package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Rootwalla — Creature — Lizard {R}, 1/1 (Mystery Booster 2):
//
//	"{R}: This creature gets +2/+0 until end of turn. Activate only once
//	 each turn.
//	 Madness {0} (If you discard this card, discard it into exile. When you
//	 do, cast it for its madness cost or put it into your graveyard.)"
//
// Basking Rootwalla's once-a-turn pump (OncePerTurnActivation, CR 602.5b) at
// {R} for +2/+0, and madness for a real {0} (CR 702.35a): a discarded
// Rootwalla is cast from exile for nothing.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +2/+0 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "d43f5977-0532-4653-bc57-ce29f9a531c9",
		Name:         "Blazing Rootwalla",
		Completeness: CompletenessFull,
		Madness:      "{0}",
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{R}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 0, "Blazing Rootwalla — +2/+0 until end of turn"),
		}},
	})
}
