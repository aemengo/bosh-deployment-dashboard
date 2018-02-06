module Status exposing (..)

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