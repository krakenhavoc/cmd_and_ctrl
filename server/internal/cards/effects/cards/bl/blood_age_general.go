package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Age General — Creature — Spirit Warrior {1}{R}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"{T}: Attacking Spirits get +1/+0 until end of turn."
//
// The set of attacking Spirits is fixed as the ability resolves (CR 611.2c): a
// Spirit that is attacking then gets +1/+0, one that starts attacking later
// does not. The General is a Spirit itself and counts when it is attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10ea73c8-7cf8-4a48-a0cc-072bb541880f",
		Name:         "Blood Age General",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Attacking Spirits get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.IsAttackingCreature(), effects.HasSubtype("Spirit")),
					Power: 1, Label: "Blood Age General — +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
