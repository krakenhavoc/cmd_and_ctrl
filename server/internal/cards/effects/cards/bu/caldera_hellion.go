package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caldera Hellion — {3}{R}{R} Creature — Hellion 3/3 (Shards of Alara, S82 OLD
// b514):
//
//	"Devour 1 (As this creature enters, you may sacrifice any number of
//	 creatures. It enters with that many +1/+1 counters on it.)
//	 When this creature enters, it deals 3 damage to each creature."
//
// The damage is dealt by the Hellion to every creature on the battlefield,
// itself included, as one damage instance. A board wipe, so it declares its
// Sweep (ADR 0126).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86a7dba7-aff9-4086-b149-03eee812249f",
		Name:         "Caldera Hellion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3}},
		Replacements: []game.ReplacementEffect{effects.Devour("Caldera Hellion", 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Caldera Hellion — it deals 3 damage to each creature", func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEachMatching(effects.NewContext(g, item), effects.Creature(), 3)
			}),
		},
	})
}
