package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlewise Aven — Creature — Bird Soldier {3}{W}, 2/2 (Judgment,
// #5611):
//
//	"Flying
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has first strike."
//
// Mystic Zealot's threshold pair with first strike for flying: the
// layer 7c self-modifier (CR 613.4c) and a layer 6 self-grant, both read
// live on every recompute, so first strike (CR 702.7b) comes and goes
// with the seventh card. Flying is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f165e482-5b87-4062-87ac-bbe8303f74c9",
		Name:            "Battlewise Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1), effects.FraThresholdSelfKeywords("first strike")},
	})
}
