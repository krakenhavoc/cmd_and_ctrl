package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thor Odinson — Legendary Creature — God Warrior Hero {3}{R}{W}, 4/4
// (Marvel Super Heroes, #4455):
//
//	"Flying, vigilance, prowess, prowess (Whenever you cast a noncreature
//	 spell, this creature gets +1/+1 until end of turn twice.)"
//
// Prowess printed twice is two instances (CR 702.108b): the keyword list
// names it twice and the engine keeps both, so one noncreature spell
// puts two triggers on the stack (Ruric Thar's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec9b74d0-e4fe-4764-b628-992a62f2a1cf",
		Name:            "Thor Odinson",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", game.KeywordProwess, game.KeywordProwess},
	})
}
