module Metric exposing (Metric, decodeMetrics)

import Json.Decode exposing (Decoder, string, int, float, list)
import Json.Decode.Pipeline exposing (decode, required, optional, hardcoded)

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