package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jwari Scuttler — Creature — Crab {2}{U}, 2/3 (Rise of the Eldrazi, S82 OLD batch 484, #5581):
// a vanilla creature with no abilities, registered so the catalog knows the
// card; its size is the printed one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d1f5f6f-af3d-4b9b-acaf-debf885f660a",
		Name:         "Jwari Scuttler",
		Completeness: effects.CompletenessFull,
	})
}
