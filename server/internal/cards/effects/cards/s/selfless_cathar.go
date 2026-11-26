package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selfless Cathar — Creature — Human Cleric {W}, 1/1 (Magic 2015, #5556):
//
//	"{1}{W}, Sacrifice this creature: Creatures you control get +1/+1 until
//	 end of turn."
//
// Encampment Keeper's shape without the tap: the creatures it affects are
// the ones its controller controls as the ability resolves (CR 611.2c), the
// Cathar already gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6e4c8de-4c35-4ea4-b865-596ae8cef075",
		Name:         "Selfless Cathar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Selfless Cathar — creatures you control get +1/+1",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
