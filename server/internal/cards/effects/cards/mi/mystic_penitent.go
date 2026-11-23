package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Penitent — Creature — Human Nomad Mystic {W}, 1/1 (Odyssey,
// #5613):
//
//	"Vigilance
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+1 and has flying."
//
// Vigilance is a printed keyword. The threshold is Mystic Zealot's pair:
// the layer 7c self-modifier (CR 613.4c) and a layer 6 flying grant,
// both read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d3d59704-569d-42a7-8b31-8a31b9c0395f",
		Name:            "Mystic Penitent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{effects.FraThresholdSelfPT(1, 1), effects.FraThresholdSelfKeywords("flying")},
	})
}
