module Msg exposing (..)

import Time
import Http
import Bootstrap.Accordion as Accordion
import Metric exposing (Metric)

type Msg
    = FetchMetrics Time.Time
    | DismissProgress
    | FilterLabels String
    | FilterDeployments String
    | Accordion Accordion.State
    | FetchMetricsResult (Result Http.Error (List Metric))