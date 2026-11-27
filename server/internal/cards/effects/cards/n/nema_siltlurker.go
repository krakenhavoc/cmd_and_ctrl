package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nema Siltlurker — Creature — Lizard {4}{G}, 3/5 (Rise of the Eldrazi, S82 OLD batch 484, #5581):
// a vanilla creature with no abilities, registered so the catalog knows the
// card; its size is the printed one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4542dff-1ba1-4b15-beaf-51a7a42782cf",
		Name:         "Nema Siltlurker",
		Completeness: effects.CompletenessFull,
	})
}
