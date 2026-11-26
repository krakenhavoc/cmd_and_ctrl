package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Otherworld Atlas — Artifact — Book {4}:
//
//	"{T}: Put a charge counter on this artifact.
//	 {T}: Each player draws a card for each charge counter on this artifact."
//
// The count is read at resolution, live while the Atlas is still there and as
// it last existed once it has gone (CR 608.2h). Each player draws that many,
// starting with the Atlas's controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d907a1b-f301-46a3-aa1c-25fb7f29c949",
		Name:         "Otherworld Atlas",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true},
			Effect:  effects.PutChargeCounterOnThis,
		}, {
			Label:   "{T}: Each player draws a card for each charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.CountersOnThisAsItResolves("charge")(ctx)
				if n <= 0 {
					return nil
				}
				for _, id := range effects.TablePlayers(ctx) {
					if err := (effects.DrawCards{Player: id, N: n}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
