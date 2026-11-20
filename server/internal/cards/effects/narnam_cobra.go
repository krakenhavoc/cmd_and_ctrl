package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narnam Cobra — Artifact Creature — Snake {2}, 2/1 (Kaladesh, #5539):
//
//	"{G}: This creature gains deathtouch until end of turn. (Any amount
//	 of damage it deals to a creature is enough to destroy it.)"
//
// Femeref Knight's self grant: a layer-6 deathtouch pinned to this object
// until the turn ends (CR 613.1f, CR 514.2) through the shared
// thisGainsKeywordUntilEndOfTurn body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ac0d584-997b-4d11-aecf-f1494eebc94c",
		Name:         "Narnam Cobra",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{G}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Narnam Cobra — deathtouch until end of turn"),
		}},
	})
}
