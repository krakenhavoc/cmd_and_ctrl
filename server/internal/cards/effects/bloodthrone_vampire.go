package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodthrone Vampire — Creature — Vampire {1}{B}, 1/1 (Rise of the Eldrazi,
// #5581):
//
//	"Sacrifice a creature: This creature gets +2/+2 until end of turn."
//
// Phyrexian Ghoul's sacrifice outlet. It can sacrifice itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa76cc50-5be0-4844-b3a4-fb33fca3cdd3",
		Name:         "Bloodthrone Vampire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Bloodthrone Vampire — +2/+2 until end of turn"),
		}},
	})
}
