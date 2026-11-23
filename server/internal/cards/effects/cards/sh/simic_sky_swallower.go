package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simic Sky Swallower — Creature — Leviathan {5}{G}{U}, 6/6 (Dissension, #5598):
//
//	"Flying, trample
//	 Shroud (This creature can't be the target of spells or abilities.)"
//
// Three printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e5c48c6d-a7bf-45b3-b48c-dc884d49c5f7",
		Name:            "Simic Sky Swallower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "shroud"},
	})
}
