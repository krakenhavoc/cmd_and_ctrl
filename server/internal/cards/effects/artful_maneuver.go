package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artful Maneuver — Instant {1}{W}:
//
//	"Target creature gets +2/+2 until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b58920e-3e90-44ed-a0e9-4cb1b7359a5b",
		Name:            "Artful Maneuver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve:       boostTheTargetUntilEOT(2, 2, "Artful Maneuver — +2/+2"),
	})
}
