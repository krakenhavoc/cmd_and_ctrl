package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Familiar — Creature — Bird {1}{W}, 1/2 (Torment, #5612):
//
//	"Flying
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has protection from black."
//
// Mystic Zealot's threshold pair with protection from black for flying:
// the layer 7c self-modifier (CR 613.4c) and a layer 6 self-grant, both
// read live on every recompute, so the protection (CR 702.16b) comes and
// goes with the seventh card. Flying is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d3931a2-baba-4e6e-b923-44df00900e51",
		Name:            "Mystic Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1), effects.FraThresholdSelfKeywords("protection from black")},
	})
}
