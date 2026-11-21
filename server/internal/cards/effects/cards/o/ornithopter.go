package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornithopter — Artifact Creature — Thopter {0}, 0/2 (EDHREC rank
// 775):
//
//	"Flying"
//
// A free flying body — the artifact deck's cheapest artifact and the
// Cathars' Crusade deck's cheapest trigger. Flying is printed, so it
// rides PrintedKeywords and the combat engine honours it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a3a98bc9-caa0-49b7-951c-fe4e4f54e4ba",
		Name:            "Ornithopter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
