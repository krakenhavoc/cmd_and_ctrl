package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Syphon — Artifact {1}{U}{U}:
//
//	"Start your engines!
//	 {2}, {T}: Draw a card.
//	 Max speed — Whenever you draw a card, each opponent mills two
//	 cards."
//
// ADR 0138 (#2122). The mill is an ordinary draw trigger, one per card
// drawn (CR 121.2), that triggers only while the Syphon's controller has
// max speed (CR 702.178a). Every other player is an opponent; one who
// has left the game is skipped. The opponents mill one after another in
// APNAP order (millEachPlayerInTurn, #4266), so two libraries that each
// ask for a replacement order open one prompt at a time.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8956775-fda8-475d-9e8a-b61196a958ea",
		Name:            "Aether Syphon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{N: 1}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(WheneverYouDraw("Aether Syphon — each opponent mills two cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return millEachPlayerInTurn(ctx, ctx.Opponents(), constantMill(2), nil)
				})),
		},
	})
}
