package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Into the Time Vortex — Sorcery {4}{R}:
//
//	"Cascade (When you cast this spell, exile cards from the top of
//	 your library until you exile a nonland card that costs less. You
//	 may cast it without paying its mana cost. Put the exiled cards on
//	 the bottom in a random order.)
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Both keywords, and nothing else. Cascade triggers on every cast, so
// the free rebound cast in the upkeep cascades again (CR 702.85a).
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ad9d969e-def5-45a3-b65b-0c776f62ef0e",
		Name:            "Into the Time Vortex",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Triggered:       []game.TriggeredAbility{Cascade()},
	})
}
