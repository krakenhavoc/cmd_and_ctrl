package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Two-Headed Hunter // Twice the Rage — Creature — Giant {4}{R}, 5/4 //
// Instant — Adventure {1}{R} (Wilds of Eldraine, #4471):
//
//	Two-Headed Hunter — "Menace"
//	Twice the Rage — "Target creature gains double strike until end of
//	 turn. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). Twice the Rage is Ticked Off's grant, declared the way
// Dual-Sun Technique declares it; any creature may be the target. The
// Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const twoHeadedHunterOracleID = "0f222f3d-f02b-42b9-aedf-fb7ed92d4889"

func init() {
	Register(Spec{
		OracleID:        twoHeadedHunterOracleID,
		Name:            "Two-Headed Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	Register(Spec{
		OracleID:     twoHeadedHunterOracleID + "#1",
		Name:         "Twice the Rage",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "double strike"))},
		Targets: TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return grantEachLegalTargetUntilEOT(ctx, "double strike", "Twice the Rage — double strike until end of turn")
		},
	})
}
