package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prakhata Pillar-Bug — Artifact Creature — Insect {3}, 2/3 (Kaladesh, #5539):
//
//	"{B}: This creature gains lifelink until end of turn. (Damage dealt
//	 by this creature also causes you to gain that much life.)"
//
// Femeref Knight's self grant: a layer-6 lifelink pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ec95969-4a6f-49f3-af23-1680276e7da2",
		Name:         "Prakhata Pillar-Bug",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gains lifelink until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("lifelink", "Prakhata Pillar-Bug — lifelink until end of turn"),
		}},
	})
}
