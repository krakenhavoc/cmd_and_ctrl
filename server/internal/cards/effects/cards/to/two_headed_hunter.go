package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const TwoHeadedHunterOracleID = "0f222f3d-f02b-42b9-aedf-fb7ed92d4889"

func init() {
	effects.Register(effects.Spec{
		OracleID:        TwoHeadedHunterOracleID,
		Name:            "Two-Headed Hunter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	effects.Register(effects.Spec{
		OracleID:     TwoHeadedHunterOracleID + "#1",
		Name:         "Twice the Rage",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{Answers: game.AnswerCombatGrant,
			Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike"))},
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantEachLegalTargetUntilEOT(ctx, "double strike", "Twice the Rage — double strike until end of turn")
		},
	})
}
