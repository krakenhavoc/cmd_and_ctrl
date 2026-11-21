package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venomthrope — Creature — Tyranid {1}{G}{U}, 2/2 (Warhammer 40,000
// Commander, S82 GAP batch 193, #5948):
//
//	"Flying, deathtouch, hexproof"
//
// Three printed keywords and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "efd70846-2226-43c6-80fe-6acdd15d4246",
		Name:            "Venomthrope",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "hexproof"},
	})
}
