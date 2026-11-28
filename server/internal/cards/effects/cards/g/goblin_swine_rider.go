package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Swine-Rider — Creature — Goblin {R}, 1/1 (Visions, S82 OLD batch
// 786, #5631):
//
//	"Whenever this creature becomes blocked, it deals 2 damage to each
//	 attacking creature and each blocking creature."
//
// "Becomes blocked" is once per blocked attacker however many creatures
// block it (CR 509.3c). The creatures hit are those attacking or blocking as
// the trigger resolves, whoever controls them and the Swine-Rider itself
// included, dealt as one instruction (damageEachMatching).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ebe32fc-3b6e-4cb9-9af0-1a6c3c6ad885",
		Name:         "Goblin Swine-Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventBecomesBlocked, effects.Self,
				"Goblin Swine-Rider — it deals 2 damage to each attacking creature and each blocking creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(effects.Creature(), effects.AttackingOrBlocking()), 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}}),
		},
	})
}
