module Subscriptions exposing (subscriptions)

import Model exposing (Model)
import Bootstrap.Accordion as Accordion
import Msg exposing (Msg)
import Time

subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch [
        (Accordion.subscriptions model.accordionState Msg.Accordion),
        (Time.every (15 * Time.second) Msg.FetchMetrics)
    ]
