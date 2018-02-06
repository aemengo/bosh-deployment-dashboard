module Service exposing (fetchMetrics)

import Msg exposing (Msg)
import Metric
import Http

fetchMetrics : Cmd Msg
fetchMetrics =
    Http.send Msg.FetchMetricsResult
        (Http.get "/api/health" Metric.decodeMetrics)