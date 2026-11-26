package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Fleet Aerialist — Creature — Human Pirate {1}{U}, 1/2 (Ixalan):
//
//	"Flying
//	 Raid — This creature enters with a +1/+1 counter on it if you attacked
//	 this turn."
//
// Rigging Runner's raid: a self-replacement on the entry (CR 614.1c), read as
// the Aerialist enters against the turn's attack record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b128d0cf-77d9-4751-99c6-0f3bc3660318",
		Name:            "Storm Fleet Aerialist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    []game.ReplacementEffect{effects.AvB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
	})
}
