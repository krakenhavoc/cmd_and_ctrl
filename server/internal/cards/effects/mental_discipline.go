package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mental Discipline — Enchantment {1}{U}{U} (Urza's Destiny, #5621):
//
//	"{1}{U}, Discard a card: Draw a card."
//
// Charging Strifeknight's rummage on an enchantment, with {1}{U} in
// place of {T}: it can be activated any number of times. The discard is
// a cost, chosen and paid at activation, so an empty hand can't pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b22080d6-a9ed-4bdd-a604-058e0e3e9463",
		Name:         "Mental Discipline",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{U}"), DiscardACard()),
			Effect:  ebDrawACard,
		}},
	})
}
