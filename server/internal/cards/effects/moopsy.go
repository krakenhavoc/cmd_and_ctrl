package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moopsy — {G} Creature — Beast, 1/1 (Star Trek, #4973):
//
//	"Deathtouch
//	 {2}: This creature gains reach until end of turn."
//
// Frog Butler's reach activation without its mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "89b9efc8-d1cd-4b2a-acf3-4f0dd1629167",
		Name:            "Moopsy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("reach", "Moopsy — reach until end of turn"),
		}},
	})
}
