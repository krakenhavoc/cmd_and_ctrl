package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Swordtooth — Creature — Dinosaur {3}{R}{G}, 5/5 (Lost Caverns of
// Ixalan Commander):
//
//	"Trample
//	 When this creature enters, it deals 1 damage to each other creature."
//
// Crater Hellion's body with 1 in place of 4: every other creature, yours
// included, takes the damage at once.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "626709bc-6c29-434f-9e3f-b84bb8a5c431",
		Name:            "Raging Swordtooth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Raging Swordtooth — 1 damage to each other creature", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.NotSelf(item.SourceCardID)), 1)
			}),
		},
	})
}
