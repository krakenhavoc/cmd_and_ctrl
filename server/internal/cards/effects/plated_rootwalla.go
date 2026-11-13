package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plated Rootwalla — Creature — Lizard {4}{G}, 3/3 (Exodus, #5626):
//
//	"{2}{G}: This creature gets +3/+3 until end of turn. Activate only
//	 once each turn."
//
// The shared thisGetsUntilEndOfTurn body behind Radha's Firebrand's
// OncePerTurnActivation (CR 602.5b): the count is of activations, so one
// that was countered still uses up the turn, and the Rootwalla may be
// activated again next turn.
//
// No simplification.
func init() {
	const pump = "{2}{G}: This creature gets +3/+3 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "e20b80f2-9913-4886-a61b-04f3fd27c2ad",
		Name:         "Plated Rootwalla",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     pump,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:      ManaCost("{2}{G}"),
			Condition: OncePerTurnActivation(pump),
			Effect:    thisGetsUntilEndOfTurn(3, 3, "Plated Rootwalla — +3/+3 until end of turn"),
		}},
	})
}
