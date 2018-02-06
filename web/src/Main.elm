module Main exposing (..)

import Html exposing (Html, text, div, input, h1, img, br, p, span)
import Html.Attributes exposing (style, class, hidden)
import List.Extra exposing (groupWhile, getAt, uniqueBy, remove)
import Bootstrap.CDN as CDN
import Bootstrap.Grid as Grid
import Bootstrap.Alert as Alert
import Bootstrap.ButtonGroup as ButtonGroup
import Bootstrap.Button as Button
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Form.Input as Input
import Bootstrap.Accordion as Accordion
import Bootstrap.Table as Table
import Bootstrap.Card as Card
import Bootstrap.Progress as Progress
import Http
import Json.Decode exposing (Decoder, map, field, string, int, float, list)
import Json.Decode.Pipeline exposing (decode, required, optional, hardcoded)
import Time
import Task
import Process

---- MODEL ----

type Status
    = Running
    | NeedsAttention

statusMessage : Status -> String
statusMessage status =
    case status of
      Running ->
        "running"
      NeedsAttention ->
        "needs attention"

type alias MetricGroup =
    { name: String
    , status: Status
    , label: String
    , vmCount: Int
    }

type alias Model =
    { metrics : List Metric
    , query : String
    , accordionState : Accordion.State
    , labelState : List String
    , isUpdating : Bool
    }

type alias Metric =
    { id : Int
    , instanceId : String
    , name : String
    , address : String
    , az : String
    , deployment : String
    , instanceIndex : Int
    , ip : String
    , label : String
    , cpuUsed : Float
    , memoryUsed : Float
    , persistentDiskUsed : Float
    , load15 : Float
    , uptime : Int
    , updatedAt : String
    , details : String
    }

blankMetric : Metric
blankMetric =
    { id = -1
    , instanceId = ""
    , name = ""
    , address = ""
    , az = ""
    , deployment = ""
    , instanceIndex = -1
    , ip = ""
    , label = ""
    , cpuUsed = 0.0
    , memoryUsed = 0.0
    , persistentDiskUsed = 0.0
    , load15 = 0.0
    , uptime = 0
    , updatedAt = ""
    , details = ""
    }

metrics : List Metric
metrics =
    [{ id = 1
    , instanceId = "1234"
    , name = "some-vm"
    , address = "api.10.0.0.1"
    , az = "z1"
    , deployment = "dedicated-mysql"
    , instanceIndex = 0
    , ip = "10.0.0.1"
    , label = "mysql"
    , cpuUsed = 10.2
    , memoryUsed = 40.5
    , persistentDiskUsed = 31.9
    , load15 = 8
    , uptime = 0
    , updatedAt = "now"
    , details = "nothing bad"
    }, { id = 2
    , instanceId = "1234"
    , name = "some-vm"
    , address = "api.10.0.0.2"
    , az = "z1"
    , deployment = "dedicated-mysql"
    , instanceIndex = 1
    , ip = "10.0.1.2"
    , label = "mysql"
    , cpuUsed = 10.5
    , memoryUsed = 48.5
    , persistentDiskUsed = 16.9
    , load15 = 15
    , uptime = 0
    , updatedAt = "yesterday"
    , details = "nothing bad over here"
    }, { id = 3
    , instanceId = "1234"
    , name = "some-vm"
    , address = "api.10.0.0.2"
    , az = "z1"
    , deployment = "cf"
    , instanceIndex = 0
    , ip = "10.0.1.2"
    , label = "cf"
    , cpuUsed = 10.5
    , memoryUsed = 48.5
    , persistentDiskUsed = 16.9
    , load15 = 15
    , uptime = 0
    , updatedAt = "yesterday"
    , details = ""
    }
    ]


init : ( Model, Cmd Msg )
init =
    ({ metrics = []
    , query = ""
    , accordionState = Accordion.initialState
    , labelState = []
    , isUpdating = False
    }, fetchMetrics )

---- UPDATE ----

type Msg
    = FetchMetrics Time.Time
    | DismissProgress
    | FilterLabelsMsg String
    | FilterDeploymentsMsg String
    | AccordionMsg Accordion.State
    | FetchMetricsResult (Result Http.Error (List Metric))

update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        FilterLabelsMsg label ->
            let
                newLabels =
                    if (List.member label model.labelState) then
                        List.Extra.remove label model.labelState
                    else
                        label :: model.labelState
            in 
                ( { model | labelState = newLabels } , Cmd.none )
        AccordionMsg state ->
            ( { model | accordionState = state }, Cmd.none )
        FilterDeploymentsMsg q ->
            ( { model | query = q }, Cmd.none )
        FetchMetrics _ ->
            ( { model | isUpdating = True }, fetchMetrics )
        DismissProgress ->
            ( { model | isUpdating = False }, Cmd.none )
        FetchMetricsResult (Ok ms) ->
            ( { model | metrics = ms }, dismissProgress )
        FetchMetricsResult (Err err) ->
            ( model, dismissProgress )
        
decodeMetric : Decoder Metric
decodeMetric =
    decode Metric
        |> required "id" int
        |> required "instance_id" string
        |> required "name" string
        |> required "address" string
        |> required "az" string
        |> required "deployment" string
        |> required "instance_index" int
        |> required "ip" string
        |> required "label" string
        |> required "cpu_used" float
        |> required "memory_used" float
        |> required "persistent_disk_used" float
        |> required "load_15" float
        |> required "uptime" int
        |> hardcoded ""
        |> optional "details" string "no details!"

decodeMetrics : Decoder (List Metric)
decodeMetrics =
    list decodeMetric
        
fetchMetrics : Cmd Msg
fetchMetrics =
    Http.send FetchMetricsResult
        (Http.get "/api/health" decodeMetrics)

dismissProgress : Cmd Msg
dismissProgress =
    Process.sleep Time.second
    |> Task.perform (always DismissProgress)

---- VIEW ----

deploymentFrom : List Metric -> MetricGroup
deploymentFrom metrics =
    let
        firstMetric =
            getAt 0 metrics

        name =
            case firstMetric of
                Nothing ->
                    "-"
                Just v ->
                    v.deployment
    
        label =
            case firstMetric of
                Nothing ->
                    "-"
                Just v ->
                    v.label
    in
        {name = name, status = (statusFor metrics), label = label, vmCount = (List.length metrics)}

stateFor : Metric -> Status
stateFor metric =
    Running

statusFor : List Metric -> Status
statusFor metrics =
    if List.all (\x -> (stateFor x) == Running) metrics then
        Running
    else
        NeedsAttention

detailsFrom : List Metric -> List (Html msg)
detailsFrom metrics =
    metrics
        |> List.filter (\x -> x.details /= "" )
        |> List.sortBy (\x -> x.instanceIndex)
        |> List.map (\x ->
                        div [] [
                            p [] [
                                span [ class "text-info" ] [ text ("Index " ++ (toString x.instanceIndex) ++ ": ") ],
                                span [ ] [ text x.details ]
                            ]
                        ]
                    )

metricsLabels : List Metric -> List String -> List (ButtonGroup.CheckboxButtonItem Msg)
metricsLabels metrics labelStates =
    metrics
        |> List.Extra.uniqueBy (\x -> x.label)
        |> List.sortBy (\x -> x.label)
        |> List.map (\x ->
                        ButtonGroup.checkboxButton (List.member x.label labelStates)
                            [ Button.primary
                            , Button.attrs
                                [ style [("margin", "0.2em")]
                                ]
                            , Button.onClick <| FilterLabelsMsg x.label
                            ] [ text x.label ] )

metricsAccordionBlock : MetricGroup -> List Metric -> (Accordion.CardBlock msg)
metricsAccordionBlock metricGroup metrics =
    let
        filteredMetrics =
            List.filter (\x -> x.deployment == metricGroup.name) metrics

        tableRows =
            filteredMetrics
                |> List.map (\x ->
                     Table.tr [] [ Table.td [] [ text x.name ]
                                 , Table.td [] [ text (x.instanceIndex |> toString)  ]
                                 , Table.td [] [ text (x |> stateFor |> statusMessage) ]
                                 , Table.td [] [ text x.ip ]
                                 , Table.td [] [ text x.az ]
                                 , Table.td [] [ text (x.cpuUsed |> toString)  ]
                                 , Table.td [] [ text (x.memoryUsed |> toString) ]
                                 , Table.td [] [ text (x.persistentDiskUsed |> toString) ]
                                 , Table.td [] [ text (x.load15 |> toString) ]
                                 ]
                 )

        table =
            Table.table
                { options = [ Table.striped, Table.small ]
                , thead = Table.simpleThead
                    [ Table.th [] [ text "name" ]
                    , Table.th [] [ text "index" ]
                    , Table.th [] [ text "status" ]
                    , Table.th [] [ text "ip" ]
                    , Table.th [] [ text "az" ]
                    , Table.th [] [ text "cpu used (%)" ]
                    , Table.th [] [ text "memory used (%)" ]
                    , Table.th [] [ text "persistent disk used (%)" ]
                    , Table.th [] [ text "load15 (%)" ]
                    ]
                , tbody =
                    Table.tbody [] tableRows
                }

        containerElement = 
            Grid.container []
                [
                    Grid.row []
                        [ Grid.col [] (detailsFrom filteredMetrics)
                        ],
                    Grid.row []
                        [ Grid.col [] [ table ]
                        ]
                ]
    in
        Accordion.block [] [
            Card.custom containerElement
        ]


deploymentAccordions :  List MetricGroup -> List Metric -> List (Accordion.Card msg)
deploymentAccordions metricGroups metrics =
    metricGroups
        |> List.sortBy (\x -> x.name)
        |> List.map (\x ->
            Accordion.card
                { id = x.name
                , options = [ Card.attrs [ style [("margin-bottom", "1.5em")] ] ]
                , header =
                    Accordion.header [] <| Accordion.toggle [] [ text (x.name ++ " - " ++ (x.status |> statusMessage) ++ " - " ++ (toString x.vmCount) ++ " VM(s)")  ]
                , blocks = [ (metricsAccordionBlock x metrics) ]
                })        

filterMetricsByQuery : String -> List Metric -> List Metric
filterMetricsByQuery query unfilteredMetrics =
    if (String.trim query) == "" then
        unfilteredMetrics
    else
        List.filter ( \x -> String.contains query x.deployment ) unfilteredMetrics

filterMetricsByLabels : List String -> List Metric -> List Metric
filterMetricsByLabels labels unfilteredMetrics =
    if List.isEmpty labels then
        unfilteredMetrics
    else
        List.filter (\x -> List.member x.label labels ) unfilteredMetrics


view : Model -> Html Msg
view { metrics, query, accordionState, labelState, isUpdating } =
    let
        filteredMetrics =
            metrics
            |> filterMetricsByQuery query
            |> filterMetricsByLabels labelState
            
        deployments =
            filteredMetrics
                |> groupWhile (\x y -> x.deployment == y.deployment )
                |> List.map (\x -> deploymentFrom x)

        deploymentsCount =
            deployments
                |> List.length
                |> toString

        brokenVMsCount =
            filteredMetrics
                |> List.filter (\x -> (stateFor x) == NeedsAttention )
                |> List.length
                
        alert =
            if (List.length filteredMetrics == 0) then
                []
            else if (brokenVMsCount == 0) then
                [ Alert.success [ text ("All " ++  deploymentsCount ++ " deployments are doing well.") ] ]
            else
                [ Alert.warning [ text ((toString brokenVMsCount) ++ " out of " ++ (toString (List.length filteredMetrics)) ++ " VMs need attention!"  ) ] ]

        body =
            if (List.length filteredMetrics) == 0 then
                [ div []
                    [ h1 [] [ text "There are no metrics to show..." ]
                    ]
                ]
            else
                [ div [ style [("margin", "0.5em")] ] [ text "Labels:" ]
                , ButtonGroup.checkboxButtonGroup [] (metricsLabels metrics labelState)
                , div [ style [("padding", "2em")] ] [ Accordion.config AccordionMsg
                    |> Accordion.withAnimation
                    |> Accordion.cards (deploymentAccordions deployments filteredMetrics)
                    |> Accordion.view accordionState
                    ]
                ]

    in
        div []
            [ CDN.stylesheet
            , Progress.progress [ Progress.attrs [ hidden (not isUpdating) ], Progress.height 6, Progress.animated, Progress.value 100 ]
            , Grid.container []
                [ Grid.row []
                    [ Grid.col []
                        [ div [ style [("padding", "1em")] ] alert
                        , div [ style [("padding", "1em")] ]
                            [ InputGroup.config
                                (InputGroup.text [ Input.placeholder "filter deployments", Input.onInput FilterDeploymentsMsg ])
                                |> InputGroup.large
                                |> InputGroup.view
                            ]
                        , div [] body
                        ]
                    ]
                ]
            ]


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch [
        (Accordion.subscriptions model.accordionState AccordionMsg),
        (Time.every (15 * Time.second) FetchMetrics)
    ]

---- PROGRAM ----


main : Program Never Model Msg
main =
    Html.program
        { view = view
        , init = init
        , update = update
        , subscriptions = subscriptions
        }
