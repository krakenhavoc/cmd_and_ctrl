package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wings of the Guard — Creature — Bird {1}{W}, 1/1 (Conspiracy: Take the
// Crown, S82 older sets batch 270, #5540):
//
//	"Flying
//	 Melee (Whenever this creature attacks, it gets +1/+1 until end of turn
//	 for each opponent you attacked this combat.)"
//
// Both keywords are the engine's: flying and melee (CR 702.121,
// game/melee.go), as on Skyhunter Strike Force.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73c0a809-fb0e-4773-a385-a6e32feb03cd",
		Name:            "Wings of the Guard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordMelee},
	})
}
