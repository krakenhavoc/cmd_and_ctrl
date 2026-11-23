package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghost Tactician — Creature — Spirit Spellshaper {4}{W}, 2/5 (Planar Chaos, #5595):
//
//	"{W}, {T}, Discard a card: Creatures you control get +1/+0 until end of
//	 turn."
//
// Castle Embereth's team pump with a discard in the cost: the creatures
// are read as the ability resolves, so one that arrives later is not
// pumped (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d89aba8-20ce-44cd-9d82-3d16fdf6aac8",
		Name:         "Ghost Tactician",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}, Discard a card: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
					Label: "Ghost Tactician — creatures you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
