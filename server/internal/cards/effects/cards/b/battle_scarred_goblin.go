package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle-Scarred Goblin — Creature — Goblin Warrior {1}{R}, 2/2 (The
// Hobbit Eternal, HOC):
//
//	"Whenever this creature becomes blocked, it deals 1 damage to each
//	 creature blocking it."
//
// Trailblazer's Torch's trigger on the creature itself, for 1. "Becomes
// blocked" is once per blocked attacker however many creatures block it
// (CR 509.3c), and "each creature blocking it" is read as the trigger
// resolves, so a blocker removed in response is not dealt damage. The
// damage is the Goblin's, dealt as one instruction to every blocker
// (#4454).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e765710-e567-4af0-b5f5-7c182649b757",
		Name:         "Battle-Scarred Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self,
				"Battle-Scarred Goblin — it deals 1 damage to each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					blockers := effects.BlockersOf(g, item.SourceCardID)
					if len(blockers) == 0 {
						return nil
					}
					return g.DealDamageEachKeyedForEffect(item.SourceCardID, blockers, 1, game.KeyedTail{})
				}),
		},
	})
}
