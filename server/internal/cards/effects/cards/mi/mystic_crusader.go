package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Crusader — Creature — Human Nomad Mystic {1}{W}{W}, 2/1
// (Odyssey, #5613):
//
//	"Protection from black and from red
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has flying."
//
// Mystic Zealot's threshold pair (the layer 7c self-modifier and the
// layer 6 self-grant, both read live) on Paladin en-Vec's two printed
// protections (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "937d663d-c9a7-4312-a4cc-fcb54faa92bc",
		Name:            "Mystic Crusader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.ProtectionFromColor("B"), game.ProtectionFromColor("R")},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1), effects.FraThresholdSelfKeywords("flying")},
	})
}
