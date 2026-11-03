package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stonecoil Serpent — Artifact Creature — Snake {X}, 0/0 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Reach, trample, protection from multicolored
//	 This creature enters with X +1/+1 counters on it."
//
// The X counters are the printed CR 614.1c entry clause (XCounters, as
// Walking Ballista's). "Protection from multicolored" is one protection
// ability (CR 702.16) whose quality is a source with two or more colours
// (CR 105.2b): game.ProtectionQualityMulticolored, so it stops damage from,
// blocking by, targeting by and attachment of such a source like any other
// protection quality.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "78124cd1-2658-43f3-a339-0755d84154d2",
		Name:                       "Stonecoil Serpent",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		PrintedKeywords:            []string{"reach", "trample", "protection from multicolored"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
	})
}
