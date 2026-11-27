package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Sigil — Artifact {1}{W}{B} (Alara Reborn, #5586):
//
//	"{T}, Sacrifice this artifact: You gain life equal to the total life lost
//	 by all players this turn. (Damage causes loss of life.)"
//
// The amount is every seat's life lost this turn, damage included (the turn
// tally Wound Reflection reads), summed as the ability resolves (CR 608.2h), so
// life lost in response counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f39fd4bd-3157-42da-9a39-94e094b5f681",
		Name:         "Tainted Sigil",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: You gain life equal to the total life lost by all players this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				total := 0
				for _, p := range g.Seats {
					if p != nil {
						total += effects.B18LifeLostThisTurn(g, p.ID)
					}
				}
				return effects.GainLife{Amount: total}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
