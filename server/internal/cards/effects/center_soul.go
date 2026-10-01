package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Center Soul — Instant {1}{W}:
//
//	"Target creature you control gains protection from the color of
//	 your choice until end of turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The colour is chosen as the spell resolves. Rebound is the engine's
// keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "686b44ec-3446-4e1f-a15f-9d8557db6d70",
		Name:            "Center Soul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature you control", YouControl()),
		OnResolve:       protectionFromAChosenColorForTheTarget("Center Soul"),
	})
}
