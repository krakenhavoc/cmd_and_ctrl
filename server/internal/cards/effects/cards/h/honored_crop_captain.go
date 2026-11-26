package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honored Crop-Captain — {R}{W} Creature — Human Warrior 3/2 (Amonkhet, S82
// old batch 246):
//
//	"Whenever this creature attacks, other attacking creatures get +1/+0 until
//	 end of turn."
//
// The attacking creatures are chosen as the trigger resolves (CR 611.2c):
// every other attacker at that moment, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd561a9a-45d2-4e9b-82df-5d4b24cca24a",
		Name:         "Honored Crop-Captain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked,
				"Honored Crop-Captain — other attacking creatures get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.AttackingCreature(), effects.OtherThan(item.SourceCardID)),
						Power: 1,
						Label: "Honored Crop-Captain",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
