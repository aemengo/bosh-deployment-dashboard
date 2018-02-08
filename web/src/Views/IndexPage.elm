module IndexPage exposing (view)

import Metric exposing (Metric)
import MetricGroup exposing (MetricGroup)
import Msg exposing (Msg)
import Model exposing (Model)
import List.Extra
import Status
import Html exposing (Html, text, div, input, h1, img, br, p, span)
import Html.Attributes exposing (style, class, hidden)
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
                            , Button.onClick <| Msg.FilterLabels x.label
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
                                 , Table.td [] [ text (x |> Status.fromMetric |> Status.message) ]
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
                    Accordion.header [] <| Accordion.toggle [] [ text (x.name ++ " - " ++ (x.status |> Status.message) ++ " - " ++ (toString x.vmCount) ++ " VM(s)")  ]
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
                |> List.sortBy .deployment
                |> List.Extra.groupWhile (\x y -> x.deployment == y.deployment )
                |> List.map (\x -> MetricGroup.from x)

        deploymentsCount =
            deployments
                |> List.length
                |> toString

        brokenVMsCount =
            filteredMetrics
                |> List.filter (\x -> (Status.fromMetric x) == Status.NeedsAttention )
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
                , div [ style [("padding", "2em")] ] [ Accordion.config Msg.Accordion
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
                                (InputGroup.text [ Input.placeholder "filter deployments", Input.onInput Msg.FilterDeployments ])
                                |> InputGroup.large
                                |> InputGroup.view
                            ]
                        , div [] body
                        ]
                    ]
                ]
            ]


