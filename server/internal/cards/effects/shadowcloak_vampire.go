package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadowcloak Vampire — Creature — Vampire {4}{B}, 4/3 (Magic 2015,
// #5556):
//
//	"Pay 2 life: This creature gains flying until end of turn."
//
// Glorifier of Dusk's first ability: a life cost (you can't pay more
// life than you have) and a turn-scoped flying on the Vampire itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "00fc773a-72aa-4b58-a17c-8ac0f17c4c85",
		Name:         "Shadowcloak Vampire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Pay 2 life: This creature gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    PayLife(2),
			Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Shadowcloak Vampire — flying until end of turn"),
		}},
	})
}
