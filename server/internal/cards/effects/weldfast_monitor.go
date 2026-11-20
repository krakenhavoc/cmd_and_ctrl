package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weldfast Monitor — Artifact Creature — Lizard {3}, 3/2 (Kaladesh, #5539):
//
//	"{R}: This creature gains menace until end of turn."
//
// Femeref Knight's self grant: a layer-6 menace pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d604aff-ba60-4064-b025-b1f3d1c0254e",
		Name:         "Weldfast Monitor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{R}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("menace", "Weldfast Monitor — menace until end of turn"),
		}},
	})
}
