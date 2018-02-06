module MetricGroup exposing (MetricGroup)

import Status exposing (Status)

type alias MetricGroup =
    { name: String
    , status: Status.Status
    , label: String
    , vmCount: Int
    }

