module Main exposing (main)

import Browser
import Html exposing (Html, a, button, div, h1, h2, h3, header, input, label, p, section, span, table, tbody, td, text, th, thead, tr)
import Html.Attributes exposing (checked, class, classList, disabled, download, href, type_, value)
import Html.Events exposing (onCheck, onClick, onInput)
import Http
import Json.Decode as Decode
import Json.Encode as Encode


type alias Site =
    { id : String
    , name : String
    , selected : Bool
    }


type alias Ticket =
    { id : String
    , name : String
    , ok : Bool
    , transactions : Int
    , filename : String
    , download : String
    , headers : List String
    , rows : List (List String)
    , truncated : Bool
    , error : String
    }


type alias ExportResult =
    { results : List Ticket
    , zipDownload : String
    }


type alias Desk =
    { sites : List Site
    , start : String
    , end : String
    , phase : Phase
    , results : Maybe ExportResult
    }


type Phase
    = Idle
    | Loading
    | Failed String


type Model
    = Booting
    | BootFailed String
    | Open Desk


type Msg
    = GotSites (Result String ( List Site, String, String ))
    | Toggle String Bool
    | SetStart String
    | SetEnd String
    | Submit
    | GotExport (Result String ExportResult)


main : Program () Model Msg
main =
    Browser.element
        { init = \_ -> ( Booting, loadSites )
        , view = view
        , update = update
        , subscriptions = \_ -> Sub.none
        }


loadSites : Cmd Msg
loadSites =
    Http.get
        { url = "/api/sites"
        , expect = expectJson GotSites bootDecoder
        }


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case ( msg, model ) of
        ( GotSites (Ok ( sites, start, end )), _ ) ->
            ( Open
                { sites = sites
                , start = start
                , end = end
                , phase = Idle
                , results = Nothing
                }
            , Cmd.none
            )

        ( GotSites (Err reason), _ ) ->
            ( BootFailed reason, Cmd.none )

        ( Toggle id checked, Open ready ) ->
            ( Open { ready | sites = setSelected id checked ready.sites }, Cmd.none )

        ( SetStart start, Open ready ) ->
            ( Open { ready | start = start }, Cmd.none )

        ( SetEnd end, Open ready ) ->
            ( Open { ready | end = end }, Cmd.none )

        ( Submit, Open ready ) ->
            if ready.phase == Loading || not (canSubmit ready) then
                ( model, Cmd.none )

            else
                ( Open { ready | phase = Loading }, export ready )

        ( GotExport (Ok result), Open ready ) ->
            ( Open { ready | phase = Idle, results = Just result }, Cmd.none )

        ( GotExport (Err reason), Open ready ) ->
            ( Open { ready | phase = Failed reason }, Cmd.none )

        _ ->
            ( model, Cmd.none )


export : Desk -> Cmd Msg
export ready =
    Http.post
        { url = "/api/export"
        , body =
            Http.jsonBody
                (Encode.object
                    [ ( "sites", Encode.list Encode.string (selectedIds ready.sites) )
                    , ( "start", Encode.string ready.start )
                    , ( "end", Encode.string ready.end )
                    ]
                )
        , expect = expectJson GotExport exportDecoder
        }


view : Model -> Html Msg
view model =
    case model of
        Booting ->
            div [ class "boot" ] [ text "Ouverture du livre…" ]

        BootFailed reason ->
            div [ class "boot" ] [ text reason ]

        Open ready ->
            div [ class "stage" ]
                [ p [ class "kicker" ] [ text "Caisses des boutiques" ]
                , h1 [] [ text "Grand livre" ]
                , p [ class "lead" ]
                    [ text "Choisissez une ou plusieurs caisses et une période. Chaque boutique reçoit son fichier Excel. Si une caisse ne répond pas, les autres continuent." ]
                , div [ class "paper" ]
                    [ section [ class "shops" ]
                        [ p [ class "section-label" ] [ text "Boutiques" ]
                        , div [] (List.map viewSite ready.sites)
                        ]
                    , section [ class "period" ]
                        [ p [ class "section-label" ] [ text "Période" ]
                        , div [ class "dates" ]
                            [ label [ class "field" ]
                                [ text "Du"
                                , input [ type_ "date", value ready.start, onInput SetStart ] []
                                ]
                            , label [ class "field" ]
                                [ text "Au"
                                , input [ type_ "date", value ready.end, onInput SetEnd ] []
                                ]
                            ]
                        , p [ class "hint" ] [ text (dateHint ready) ]
                        , button
                            [ class "extract"
                            , type_ "button"
                            , onClick Submit
                            , disabled (ready.phase == Loading || not (canSubmit ready))
                            ]
                            [ text (submitLabel ready) ]
                        , p [ class "note" ]
                            [ text "Chaque caisse produit un fichier Excel à télécharger. Une caisse muette est simplement ignorée." ]
                        ]
                    ]
                , viewStatus ready
                , viewResults ready.results
                ]


viewSite : Site -> Html Msg
viewSite site =
    label [ classList [ ( "shop", True ), ( "shop--on", site.selected ) ] ]
        [ input
            [ type_ "checkbox"
            , checked site.selected
            , onCheck (Toggle site.id)
            ]
            []
        , span [ class "shop__name" ] [ text site.name ]
        ]


viewStatus : Desk -> Html Msg
viewStatus ready =
    case ready.phase of
        Failed reason ->
            div [ class "banner" ] [ text reason ]

        _ ->
            text ""


viewResults : Maybe ExportResult -> Html Msg
viewResults maybe =
    case maybe of
        Nothing ->
            text ""

        Just result ->
            section [ class "results" ]
                [ div [ class "summary" ]
                    [ h2 [] [ text (summary result) ]
                    , viewZip result.zipDownload
                    ]
                , div [] (List.indexedMap viewTicket result.results)
                ]


viewZip : String -> Html Msg
viewZip url =
    if url == "" then
        text ""

    else
        a [ class "zip", href url ] [ text "Tout télécharger (.zip)" ]


viewTicket : Int -> Ticket -> Html Msg
viewTicket index ticket =
    articleDelay index
        [ header []
            [ h3 [] [ text ticket.name ]
            , span [ class (stampClass ticket) ] [ text (stamp ticket) ]
            ]
        , p [] [ text (ticketText ticket) ]
        , viewDownload ticket
        , viewTable ticket
        ]


articleDelay : Int -> List (Html Msg) -> Html Msg
articleDelay _ children =
    div [ class "ticket" ] children


viewDownload : Ticket -> Html Msg
viewDownload ticket =
    if ticket.download == "" then
        text ""

    else
        a [ class "download", href ticket.download, download ticket.filename ]
            [ text "Télécharger .xlsx" ]


viewTable : Ticket -> Html Msg
viewTable ticket =
    if not ticket.ok || List.isEmpty ticket.headers then
        text ""

    else
        div [ class "table-wrap" ]
            [ table []
                [ thead []
                    [ tr [] (List.map (\name -> th [] [ text name ]) ticket.headers) ]
                , tbody [] (List.map viewRow ticket.rows)
                ]
            , if ticket.truncated then
                p [] [ text "Aperçu limité aux premières lignes. Le fichier Excel contient tout." ]

              else
                text ""
            ]


viewRow : List String -> Html Msg
viewRow cells =
    tr [] (List.map (\cell -> td [] [ text cell ]) cells)


stamp : Ticket -> String
stamp ticket =
    if ticket.ok then
        "Prête"

    else
        "Ignorée"


stampClass : Ticket -> String
stampClass ticket =
    if ticket.ok then
        "stamp-ok"

    else
        "stamp-no"


ticketText : Ticket -> String
ticketText ticket =
    if ticket.ok && ticket.transactions == 0 then
        "Aucune transaction sur cette période."

    else if ticket.ok then
        String.fromInt ticket.transactions ++ " transaction" ++ plural ticket.transactions ++ "."

    else
        "Ignorée : " ++ ticket.error


summary : ExportResult -> String
summary result =
    let
        ready =
            List.length (List.filter .ok result.results)

        skipped =
            List.length result.results - ready
    in
    String.fromInt ready
        ++ (if ready == 1 then
                " fichier prêt, "

            else
                " fichiers prêts, "
           )
        ++ String.fromInt skipped
        ++ (if skipped == 1 then
                " caisse ignorée."

            else
                " caisses ignorées."
           )


plural : Int -> String
plural count =
    if count == 1 then
        ""

    else
        "s"


submitLabel : Desk -> String
submitLabel ready =
    if ready.phase == Loading then
        "Lecture des caisses…"

    else
        let
            count =
                List.length (selectedIds ready.sites)
        in
        if count == 0 then
            "Choisir une caisse"

        else if count == 1 then
            "Extraire 1 caisse"

        else
            "Extraire " ++ String.fromInt count ++ " caisses"


dateHint : Desk -> String
dateHint ready =
    if ready.start == "" || ready.end == "" then
        "Indiquez les deux dates."

    else if ready.start > ready.end then
        "La date de début doit précéder la date de fin."

    else
        ""


canSubmit : Desk -> Bool
canSubmit ready =
    not (List.isEmpty (selectedIds ready.sites))
        && ready.start /= ""
        && ready.end /= ""
        && ready.start <= ready.end


selectedIds : List Site -> List String
selectedIds sites =
    sites
        |> List.filter .selected
        |> List.map .id


setSelected : String -> Bool -> List Site -> List Site
setSelected id checked sites =
    List.map
        (\site ->
            if site.id == id then
                { site | selected = checked }

            else
                site
        )
        sites


bootDecoder : Decode.Decoder ( List Site, String, String )
bootDecoder =
    Decode.map3 (\sites start end -> ( sites, start, end ))
        (Decode.field "sites" (Decode.list siteDecoder))
        (Decode.field "defaultStart" Decode.string)
        (Decode.field "defaultEnd" Decode.string)


siteDecoder : Decode.Decoder Site
siteDecoder =
    Decode.map2 (\id name -> Site id name True)
        (Decode.field "id" Decode.string)
        (Decode.field "name" Decode.string)


exportDecoder : Decode.Decoder ExportResult
exportDecoder =
    Decode.map2 ExportResult
        (Decode.field "results" (Decode.list ticketDecoder))
        (Decode.field "zipDownload" Decode.string)


ticketDecoder : Decode.Decoder Ticket
ticketDecoder =
    Decode.map8 ticketFromParts
        (Decode.field "id" Decode.string)
        (Decode.field "name" Decode.string)
        (Decode.field "ok" Decode.bool)
        (Decode.field "transactions" Decode.int)
        (Decode.field "filename" Decode.string)
        (Decode.field "download" Decode.string)
        (Decode.field "headers" (Decode.list Decode.string))
        (Decode.map3 (\rows truncated error -> ( rows, truncated, error ))
            (Decode.field "rows" (Decode.list (Decode.list Decode.string)))
            (Decode.field "truncated" Decode.bool)
            (Decode.field "error" Decode.string)
        )


ticketFromParts : String -> String -> Bool -> Int -> String -> String -> List String -> ( List (List String), Bool, String ) -> Ticket
ticketFromParts id name ok transactions filename downloadUrl headers ( rows, truncated, error ) =
    Ticket id name ok transactions filename downloadUrl headers rows truncated error


expectJson : (Result String a -> Msg) -> Decode.Decoder a -> Http.Expect Msg
expectJson toMsg decoder =
    Http.expectStringResponse toMsg <|
        \response ->
            case response of
                Http.GoodStatus_ _ body ->
                    case Decode.decodeString decoder body of
                        Ok value ->
                            Ok value

                        Err _ ->
                            Err "Réponse du serveur illisible."

                Http.BadStatus_ _ body ->
                    Err (errorMessage body)

                _ ->
                    Err "La liaison avec le serveur a échoué."


errorMessage : String -> String
errorMessage body =
    case Decode.decodeString (Decode.field "error" Decode.string) body of
        Ok message ->
            message

        Err _ ->
            "La demande a échoué."
