module Main exposing (..)

import IndexPage
import Html
import Msg exposing (Msg)
import Model exposing (Model)
import Subscriptions

---- PROGRAM ----

main : Program Never Model Msg
main =
    Html.program
        { view = IndexPage.view
        , init = Model.init
        , update = Model.update
        , subscriptions = Subscriptions.subscriptions
        }
