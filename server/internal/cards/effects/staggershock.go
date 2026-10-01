package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Staggershock — Instant {2}{R}:
//
//	"Staggershock deals 2 damage to any target.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854): the
// declaration is the whole of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "056c3b7d-b603-40b8-8404-18c2eb7e7129",
		Name:            "Staggershock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetAny(),
		OnResolve:       damageToFirstTarget(2),
	})
}
