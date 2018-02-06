module Status exposing (..)

import Metric exposing (Metric)

type Status
    = Running
    | NeedsAttention

message : Status -> String
message status =
    case status of
      Running ->
        "running"
      NeedsAttention ->
        "needs attention"

fromMetric : Metric -> Status
fromMetric metric =
    Running

fromMetrics : List Metric -> Status
fromMetrics metrics =
    if List.all (\x -> (fromMetric x) == Running) metrics then
        Running
    else
        NeedsAttention