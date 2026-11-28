package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pincher Beetles — Creature — Insect {2}{G}, 3/1 (Tempest, S82 OLD batch 767,
// #5628):
//
//	"Shroud (This creature can't be the target of spells or abilities.)"
//
// A printed keyword, read through the layered view like any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0fae1c73-fd34-433e-8989-fbd80e3b1c70",
		Name:            "Pincher Beetles",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
