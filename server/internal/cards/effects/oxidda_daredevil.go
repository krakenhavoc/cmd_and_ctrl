package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oxidda Daredevil — Creature — Goblin Artificer {1}{R}, 2/1 (Scars of Mirrodin, #5577):
//
//	"Sacrifice an artifact: This creature gains haste until end of turn."
//
// An artifact sacrifice for a cost (Atog's) and a layer 6 grant pinned to
// this object (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4be254e7-d082-4c10-8495-1ef8b80721e1",
		Name:         "Oxidda Daredevil",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice an artifact: This creature gains haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"haste"}}},
			Cost:    SacrificeN(1, "an artifact", Artifact()),
			Effect:  thisGainsKeywordUntilEndOfTurn("haste", "Oxidda Daredevil — haste until end of turn"),
		}},
	})
}
