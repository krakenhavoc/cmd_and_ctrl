package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Akrasan Squire — Creature — Human Soldier {W}, 1/1 (Shards of Alara,
// #5588):
//
//	"Exalted (Whenever a creature you control attacks alone, that
//	 creature gets +1/+1 until end of turn.)"
//
// Exalted is the canonical keyword (CR 702.83a, game/exalted.go): the
// engine derives the trigger from the printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2b8ef83b-1a1d-4299-ab69-e9e296bb224d",
		Name:            "Akrasan Squire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordExalted},
	})
}
