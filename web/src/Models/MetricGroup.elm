module MetricGroup exposing (MetricGroup, from)

import Status exposing (Status)
import Metric exposing (Metric)
import List.Extra

type alias MetricGroup =
    { name: String
    , status: Status.Status
    , label: String
    , vmCount: Int
    }

from : List Metric -> MetricGroup
from metrics =
    let
        firstMetric =
            List.Extra.getAt 0 metrics

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
        {name = name, status = (Status.fromMetrics metrics), label = label, vmCount = (List.length metrics)}