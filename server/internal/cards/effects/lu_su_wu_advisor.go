package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lu Su, Wu Advisor — Legendary Creature — Human Advisor {3}{U}{U}, 1/2 (#4062):
//
//	"{T}: Draw a card.
//	 Activate only during your turn, before attackers are declared."
//
// The window is DuringYourTurnBeforeAttackersAreDeclared, an activation
// condition (CR 602.1b): your own turn, and only until the first declare
// attackers step of the turn begins - the card's ruling reads it from the
// first combat, so a turn with several combat phases does not open it
// again. Not a sorcery-speed gate: it opens your beginning, combat and
// main-phase steps up to that point, with or without an item on the stack,
// and the view and the legal enumerator read the same closure, so the
// ability greys out when the window shuts.
//
// A repeatable draw, once a turn, before your attackers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59238ba1-fde4-4e43-9f24-b24125e59de0",
		Name:         "Lu Su, Wu Advisor",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Draw a card. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Purpose:   game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    drawOneCardEffect,
		}},
	})
}
