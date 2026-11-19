package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dross Hopper — Creature — Phyrexian Insect Horror {1}{B}, 2/1:
//
//	"Sacrifice a creature: This creature gains flying until end of turn."
func init() {
	Register(Spec{
		OracleID:     "2ec8c540-73bf-4412-a53a-499e9b9fd213",
		Name:         "Dross Hopper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant | game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Dross Hopper — flying until end of turn"),
		}},
	})
}
