package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Torches — Creature — Wall {1}{R}, 4/1 (Magic 2012, #5573):
//
//	"Defender (This creature can't attack.)"
//
// Defender is a printed keyword (Glacial Wall's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "674a269c-61ed-4fb5-ad74-422e581a5a91",
		Name:            "Wall of Torches",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
