package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lagonna-Band Trailblazer — Creature — Centaur Scout {W}, 0/4 (Journey into Nyx, #5558):
//
//	"Heroic — Whenever you cast a spell that targets this creature, put
//	 a +1/+1 counter on it."
//
// Satyr Hoplite's heroic counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09cc0faa-71d7-428b-9fc4-5ba09e244e89",
		Name:         "Lagonna-Band Trailblazer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Heroic("Lagonna-Band Trailblazer — a +1/+1 counter on it (heroic)", effects.ThisGetsAPlusOneCounter),
		},
	})
}
