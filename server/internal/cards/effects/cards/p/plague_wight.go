package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plague Wight — Creature — Zombie {1}{B}, 2/1 (Ravnica Allegiance):
//
//	"Whenever this creature becomes blocked, each creature blocking it gets
//	 -1/-1 until end of turn."
//
// Baneblade Scoundrel's trigger: one per blocked event, shrinking every
// blocker as the trigger resolves.
//
// No simplification.
func init() {
	const label = "Plague Wight — each creature blocking it gets -1/-1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "b7639d86-4c7c-423f-8b13-fde3086d3cbb",
		Name:         "Plague Wight",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, effects.Self, label, effects.OldB180ShrinkEachBlocker(label)),
		},
	})
}
