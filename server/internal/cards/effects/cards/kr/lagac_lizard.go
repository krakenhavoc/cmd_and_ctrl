package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lagac Lizard — Creature — Lizard {3}{R}, 3/3 (Rise of the Eldrazi, S82 OLD batch 484, #5581):
// a vanilla creature with no abilities, registered so the catalog knows the
// card; its size is the printed one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e241c8c3-359d-4e20-97a6-a873c50f1ef9",
		Name:         "Lagac Lizard",
		Completeness: effects.CompletenessFull,
	})
}
