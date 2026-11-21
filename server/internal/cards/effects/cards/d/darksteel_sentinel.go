package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Sentinel — Artifact Creature — Golem {6}, 3/3 (Scars of
// Mirrodin, #5577):
//
//	"Flash
//	 Vigilance
//	 Indestructible"
//
// A keyword body: all three are the engine's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b91c8946-591b-4c0d-a37e-36803df40db7",
		Name:            "Darksteel Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "vigilance", "indestructible"},
	})
}
