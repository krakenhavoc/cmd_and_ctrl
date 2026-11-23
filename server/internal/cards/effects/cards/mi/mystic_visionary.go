package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mystic Visionary — Creature — Human Nomad Mystic {1}{W}, 2/1 (Odyssey,
// #5613):
//
//	"Threshold — This creature has flying as long as there are seven or
//	 more cards in your graveyard."
//
// A layer 6 self-grant of flying, read live on every recompute (the
// threshold half of Mystic Zealot without the +1/+1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bde74bcc-41b8-4014-b28e-2f6326e20799",
		Name:         "Mystic Visionary",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FraThresholdSelfKeywords("flying")},
	})
}
