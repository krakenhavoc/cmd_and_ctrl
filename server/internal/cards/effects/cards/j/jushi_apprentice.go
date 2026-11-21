package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
const JushiApprenticeOracleID = "2b565791-1b94-4dad-a48d-be7fc2198e19"

func init() {
	effects.Register(effects.Spec{
		OracleID:     JushiApprenticeOracleID,
		Name:         "Jushi Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}: Draw a card. If you have nine or more cards in hand, flip this creature.",
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := (effects.DrawCards{N: 1}).Apply(effects.NewContext(g, item)); err != nil {
					return err
				}
				return effects.FlipThisWhen(g, item, func(g *game.Game, controller uuid.UUID) bool {
					return effects.B14HandSize(g, controller) >= 9
				})
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     JushiApprenticeOracleID + "#1",
		Name:         "Tomoya the Revealer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}{U}, {T}: Target player draws X cards, where X is the number of cards in your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}{U}{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					return effects.DrawCards{Player: t.ID, N: effects.B14HandSize(g, item.Controller)}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
