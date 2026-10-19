package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Norwood Priestess — Creature — Elf Druid {2}{G}{G}, 1/1 (#4062):
//
//	"{T}: You may put a green creature card from your hand onto the battlefield.
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
// A hand with no green creature card makes the put a no-op (the player is
// not asked).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "902a7588-602c-456f-8250-c94f35e56d3a",
		Name:         "Norwood Priestess",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: You may put a green creature card from your hand onto the battlefield. Activate only during your turn, before attackers are declared.",
			Cost:      TapCost(),
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Condition: DuringYourTurnBeforeAttackersAreDeclared(),
			Effect:    norwoodPriestessPutsOne,
		}},
	})
}
