package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talas Researcher — Creature — Human Pirate Wizard {4}{U}, 1/1 (#4062):
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
		OracleID:     "1ed662d7-65b9-4024-a892-c4559fb6aaf9",
		Name:         "Talas Researcher",
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
