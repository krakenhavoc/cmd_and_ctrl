package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Bairn — Creature — Vampire {2}{B}, 2/2 (Magic 2014, #5563):
//
//	"Sacrifice another creature: This creature gets +2/+2 until end of turn."
//
// Sepulcher Ghoul's sacrifice pump without the once-a-turn limit: it can be
// activated as often as there are other creatures to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33bea2f0-c5aa-44b5-8714-5f50f7f2bb05",
		Name:         "Blood Bairn",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Effect:  thisGetsUntilEndOfTurn(2, 2, "Blood Bairn — +2/+2"),
		}},
	})
}
