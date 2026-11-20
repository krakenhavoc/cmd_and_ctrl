package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reconstructed Thopter — Artifact Creature — Thopter {3}, 2/1:
//
//	"Flying
//	 Unearth {2}"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c87ff8e8-a3da-4848-9881-fbc896acf7c5",
		Name:            "Reconstructed Thopter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{2}")},
	})
}
