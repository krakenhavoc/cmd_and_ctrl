package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Antagonism — Enchantment {3}{R} (Urza's Saga, S82 OLD b723):
//
//	At the beginning of each player's end step, this enchantment deals 2 damage to that player unless one of their opponents was dealt damage this turn.
//
// One trigger at every end step, on the active player; the damage is skipped when any other player was dealt damage this turn (damage dealt, not prevented).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f81fcf-2dcf-41b9-ad7c-588149feb12a",
		Name:         "Antagonism",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Antagonism — 2 damage to that player unless an opponent was dealt damage this turn", effects.S82OldB723Antagonism),
		},
	})
}
