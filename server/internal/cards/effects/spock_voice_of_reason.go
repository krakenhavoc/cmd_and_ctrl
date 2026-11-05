package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spock, Voice of Reason — {2}{U} Legendary Creature — Vulcan Human
// Scientist 3/2 (Star Trek Commander, #4974):
//
//	"{2}, {T}: Draw a card for each nontoken permanent you control that
//	 entered this turn."
//
// The count is read as the ability resolves (CR 608.2h), from the
// per-turn entry tally (Game.EnteredThisTurn): nontoken permanents you
// control then, that entered the battlefield this turn. One that entered
// and left is not counted; Spock counts himself if he entered this turn
// (and somehow lost summoning sickness). The {T} is the symbol, so he
// can't activate it the turn he arrives without haste (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d338db9-e968-49ec-8959-141480b17b8f",
		Name:         "Spock, Voice of Reason",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card for each nontoken permanent you control that entered this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := trekB46NontokenPermanentsYouControlThatEnteredThisTurn(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return DrawCards{N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
