package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trained Armodon — Creature — Elephant {1}{G}{G}, 3/3 (Eighth Edition, #5607). No rules text.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c9f5359-249e-488b-9a7e-db1a43c2b1f4",
		Name:         "Trained Armodon",
		Completeness: effects.CompletenessFull,
	})
}
