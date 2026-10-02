package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taigam's Strike — Sorcery {3}{U}:
//
//	"Target creature gets +2/+0 until end of turn and can't be blocked
//	 this turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f256c828-01d2-4077-823d-8dccd15120bd",
		Name:            "Taigam's Strike",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve:       boostTheTargetAndUnblockableThisTurn(2, "Taigam's Strike"),
	})
}
