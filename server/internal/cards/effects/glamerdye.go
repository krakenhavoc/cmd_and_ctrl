package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glamerdye — Instant {1}{U}:
//
//	"Change the text of target spell or permanent by replacing all
//	 instances of one color word with another.
//	 Retrace"
//
// Sleight of Mind with retrace (CR 702.81, Call the Skybreaker's
// shape): castable from the graveyard for its printed cost plus a land
// card discarded from hand. The text change is CR 612.2 (text_words.go,
// ADR 0160's 2026-10-15 amendment) and lasts indefinitely.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "570a5e60-ed0f-4c14-8cc3-af819648030b",
		Name:             "Glamerdye",
		Completeness:     CompletenessFull,
		Targets:          TargetSpellOrPermanent("target spell or permanent", nil, nil),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Retrace("{1}{U}")},
		OnResolve:        changeTextOfTarget(textColorWords, false, nothingAfterTheChange),
	})
}
