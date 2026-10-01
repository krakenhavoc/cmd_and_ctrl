package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distortion Strike — Sorcery {U}:
//
//	"Target creature gets +1/+0 until end of turn and can't be blocked
//	 this turn.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// Rebound is the engine's keyword (game/rebound.go, #1854). The
// rebound cast happens in the upkeep, so the creature is unblockable
// for that whole turn's combat.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c7f3fbf-6b68-492f-893d-853c73cfa463",
		Name:            "Distortion Strike",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Targets:         TargetCreature("target creature"),
		OnResolve:       boostTheTargetAndUnblockableThisTurn(1, "Distortion Strike"),
	})
}
