module Dashboard.Footer exposing (showFooter, tick, toggleHelp, view)

import Concourse.Cli as Cli
import Concourse.PipelineStatus as PipelineStatus exposing (PipelineStatus(..))
import Dashboard.Models exposing (FooterModel, Group)
import Dashboard.Styles as Styles
import Html exposing (Html)
import Html.Attributes exposing (attribute, class, href, id, style)
import Html.Events exposing (onMouseEnter, onMouseLeave)
import Message.Message exposing (Hoverable(..), Message(..))
import Routes
import ScreenSize


showFooter : FooterModel r -> FooterModel r
showFooter model =
    { model | hideFooter = False, hideFooterCounter = 0 }


tick : FooterModel r -> FooterModel r
tick model =
    if model.hideFooterCounter > 4 then
        { model | hideFooter = True }

    else
        { model | hideFooterCounter = model.hideFooterCounter + 1 }


toggleHelp : FooterModel r -> FooterModel r
toggleHelp model =
    { model | showHelp = not (hideHelp model || model.showHelp) }


hideHelp : { a | groups : List Group } -> Bool
hideHelp { groups } =
    List.isEmpty (groups |> List.concatMap .pipelines)


view : FooterModel r -> List (Html Message)
view model =
    if model.showHelp then
        [ keyboardHelp ]

    else if not model.hideFooter then
        [ infoBar model ]

    else
        []


keyboardHelp : Html Message
keyboardHelp =
    Html.div
        [ class "keyboard-help" ]
        [ Html.div
            [ class "help-title" ]
            [ Html.text "keyboard shortcuts" ]
        , Html.div
            [ class "help-line" ]
            [ Html.div
                [ class "keys" ]
                [ Html.span
                    [ class "key" ]
                    [ Html.text "/" ]
                ]
            , Html.text "search"
            ]
        , Html.div
            [ class "help-line" ]
            [ Html.div
                [ class "keys" ]
                [ Html.span
                    [ class "key" ]
                    [ Html.text "?" ]
                ]
            , Html.text "hide/show help"
            ]
        ]


infoBar :
    { a
        | hovered : Maybe Hoverable
        , screenSize : ScreenSize.ScreenSize
        , version : String
        , highDensity : Bool
        , groups : List Group
    }
    -> Html Message
infoBar model =
    Html.div
        [ id "dashboard-info"
        , style <|
            Styles.infoBar
                { hideLegend = hideLegend model
                , screenSize = model.screenSize
                }
        ]
    <|
        legend model
            ++ concourseInfo model


legend :
    { a
        | groups : List Group
        , screenSize : ScreenSize.ScreenSize
        , highDensity : Bool
    }
    -> List (Html Message)
legend model =
    if hideLegend model then
        []

    else
        [ Html.div
            [ id "legend"
            , style Styles.legend
            ]
          <|
            List.map legendItem
                [ PipelineStatusPending False
                , PipelineStatusPaused
                ]
                ++ [ Html.div [ style Styles.legendItem ]
                        [ Html.div [ style Styles.runningLegendItem ] []
                        , Html.div [ style [ ( "width", "10px" ) ] ] []
                        , Html.text "running"
                        ]
                   ]
                ++ List.map legendItem
                    [ PipelineStatusFailed PipelineStatus.Running
                    , PipelineStatusErrored PipelineStatus.Running
                    , PipelineStatusAborted PipelineStatus.Running
                    , PipelineStatusSucceeded PipelineStatus.Running
                    ]
                ++ legendSeparator model.screenSize
                ++ [ toggleView model.highDensity ]
        ]


concourseInfo :
    { a | version : String, hovered : Maybe Hoverable }
    -> List (Html Message)
concourseInfo { version, hovered } =
    [ Html.div [ id "concourse-info", style Styles.info ]
        [ Html.div [ style Styles.infoItem ]
            [ Html.text <| "version: v" ++ version ]
        , Html.div [ style Styles.infoItem ] <|
            [ Html.span
                [ style [ ( "margin-right", "10px" ) ] ]
                [ Html.text "cli: " ]
            ]
                ++ List.map (cliIcon hovered) Cli.clis
        ]
    ]


hideLegend : { a | groups : List Group } -> Bool
hideLegend { groups } =
    List.isEmpty (groups |> List.concatMap .pipelines)


legendItem : PipelineStatus -> Html Message
legendItem status =
    Html.div [ style Styles.legendItem ]
        [ Html.div
            [ style <| Styles.pipelineStatusIcon status ]
            []
        , Html.div [ style [ ( "width", "10px" ) ] ] []
        , Html.text <| PipelineStatus.show status
        ]


toggleView : Bool -> Html Message
toggleView highDensity =
    Html.a
        [ style Styles.highDensityToggle
        , href <| Routes.toString <| Routes.dashboardRoute (not highDensity)
        , attribute "aria-label" "Toggle high-density view"
        ]
        [ Html.div [ style <| Styles.highDensityIcon highDensity ] []
        , Html.text "high-density"
        ]


legendSeparator : ScreenSize.ScreenSize -> List (Html Message)
legendSeparator screenSize =
    case screenSize of
        ScreenSize.Mobile ->
            []

        ScreenSize.Desktop ->
            [ Html.div
                [ style Styles.legendSeparator ]
                [ Html.text "|" ]
            ]

        ScreenSize.BigDesktop ->
            [ Html.div
                [ style Styles.legendSeparator ]
                [ Html.text "|" ]
            ]


cliIcon : Maybe Hoverable -> Cli.Cli -> Html Message
cliIcon hovered cli =
    Html.a
        [ href (Cli.downloadUrl cli)
        , attribute "aria-label" <| Cli.label cli
        , style <|
            Styles.infoCliIcon
                { hovered = hovered == (Just <| FooterCliIcon cli)
                , cli = cli
                }
        , id <| "cli-" ++ Cli.id cli
        , onMouseEnter <| Hover <| Just <| FooterCliIcon cli
        , onMouseLeave <| Hover Nothing
        ]
        []