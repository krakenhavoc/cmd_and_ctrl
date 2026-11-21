package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ethercaste Knight — Artifact Creature — Human Knight {W}{U}, 1/3 (Alara
// Reborn, #5586):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Exalted is the canonical keyword (CR 702.83a, game/exalted.go): the
// engine derives the trigger from the printed keyword (Akrasan Squire).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b34a7a9e-d740-4ddc-9b3c-8c49555ee2aa",
		Name:            "Ethercaste Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
