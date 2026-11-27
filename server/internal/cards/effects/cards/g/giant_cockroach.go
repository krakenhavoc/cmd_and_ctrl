package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Cockroach — Creature — Insect {3}{B}, 4/2 (Eighth Edition, #5607). No rules text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8b49b28-d364-4cf0-a2a5-56f6f7dc1b22",
		Name:         "Giant Cockroach",
		Completeness: effects.CompletenessFull,
	})
}
