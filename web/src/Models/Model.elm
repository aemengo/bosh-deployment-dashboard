module Model exposing (Model, init, update)

import Bootstrap.Accordion as Accordion
import Metric exposing (Metric)
import Msg exposing (..)
import Service
import Process
import Time
import Task
import List.Extra

type alias Model =
    { metrics : List Metric
    , query : String
    , accordionState : Accordion.State
    , labelState : List String
    , isUpdating : Bool
    }

init : ( Model, Cmd Msg )
init =
    ({ metrics = []
    , query = ""
    , accordionState = Accordion.initialState
    , labelState = []
    , isUpdating = False
    }, Service.fetchMetrics )

update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Accordion state ->
            ( { model | accordionState = state }, Cmd.none )
        FilterLabels label ->
            ( { model | labelState = (toggleLabelState label model.labelState) } , Cmd.none )
        FilterDeployments q ->
            ( { model | query = q }, Cmd.none )
        FetchMetrics _ ->
            ( { model | isUpdating = True }, Service.fetchMetrics )
        FetchMetricsResult (Ok ms) ->
            ( { model | metrics = ms }, dismissProgress )
        FetchMetricsResult (Err err) ->
            ( model, dismissProgress )
        DismissProgress ->
            ( { model | isUpdating = False }, Cmd.none )

toggleLabelState : String -> List String -> List String
toggleLabelState newLabel labels =
    if (List.member newLabel labels) then
        List.Extra.remove newLabel labels
    else
        newLabel :: labels

dismissProgress : Cmd Msg
dismissProgress =
    Process.sleep Time.second
    |> Task.perform (always DismissProgress)