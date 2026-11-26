package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Menagerie Liberator — Creature — Human Warrior {3}{G}, 3/2 (Conspiracy:
// Take the Crown, S82 older sets batch 270, #5540):
//
//	"Trample
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)"
//
// Both keywords are the engine's: trample and melee (CR 702.121,
// game/melee.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d91dd07a-93e1-4f88-b5f1-faa56df70ed0",
		Name:            "Menagerie Liberator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordMelee},
	})
}
