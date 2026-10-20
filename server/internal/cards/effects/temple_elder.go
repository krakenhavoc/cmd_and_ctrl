package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temple Elder — Creature — Human Cleric {2}{W}, 1/2 (#4062):
//
//	"{T}: You gain 1 life.
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
// No target: the window is the only thing between it and an ordinary tap
// for life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a0ace0db-1bd6-4b30-b0e2-d266967707d7",
		Name:         "Temple Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: You gain 1 life. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    gainOneLifeEffect,
		}},
	})
}
