package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steward of Solidarity — Creature — Human Warrior {1}{W}, 2/2:
//
//	"{T}, Exert this creature: Create a 1/1 white Warrior creature token
//	 with vigilance. (An exerted creature won't untap during your next
//	 untap step.)"
//
// ADR 0130 §4: an exert cost component (ExertThis), paid with the {T}
// at announce (CR 602.2b). The Steward won't untap during your next
// untap step (CR 701.43a). A {T} ability of a creature, so it can't be
// activated the turn the Steward arrives (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dcaf3c6f-06e2-4762-82aa-625113e375a1",
		Name:         "Steward of Solidarity",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Exert this creature: Create a 1/1 white Warrior creature token with vigilance.",
			Cost:    Plus(TapCost(), ExertThis()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("1/1 white Warrior with vigilance"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
