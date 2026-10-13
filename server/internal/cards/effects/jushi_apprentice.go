package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jushi Apprentice // Tomoya the Revealer — {1}{U} Creature — Human
// Wizard 1/2 // Legendary Creature — Human Wizard 2/3, a flip card
// (#3216, ADR 0151):
//
//	Top:     "{2}{U}, {T}: Draw a card. If you have nine or more cards
//	          in hand, flip this creature."
//	Flipped: "{3}{U}{U}, {T}: Target player draws X cards, where X is
//	          the number of cards in your hand."
//
// The hand is counted after the draw. Tomoya's X is the CONTROLLER's
// hand, counted as the ability resolves, whoever the target is. Tomoya
// is blue and {1}{U} (CR 710.1c). The draw and the flip check are one
// resolution, in printed order.
//
// No simplification.
const jushiApprenticeOracleID = "2b565791-1b94-4dad-a48d-be7fc2198e19"

func init() {
	Register(Spec{
		OracleID:     jushiApprenticeOracleID,
		Name:         "Jushi Apprentice",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}, {T}: Draw a card. If you have nine or more cards in hand, flip this creature.",
			Cost:    Plus(ManaCost("{2}{U}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (DrawCards{N: 1}).Apply(NewContext(g, item)); err != nil {
					return err
				}
				return flipThisWhen(g, item, func(g *game.Game, controller uuid.UUID) bool {
					return b14HandSize(g, controller) >= 9
				})
			},
		}},
	})
	Register(Spec{
		OracleID:     jushiApprenticeOracleID + "#1",
		Name:         "Tomoya the Revealer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}{U}, {T}: Target player draws X cards, where X is the number of cards in your hand.",
			Cost:    Plus(ManaCost("{3}{U}{U}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return DrawCards{Player: t.ID, N: b14HandSize(g, item.Controller)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
