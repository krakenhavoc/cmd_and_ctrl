package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basking Rootwalla — Creature — Lizard {G}, 1/1 (Torment, #5612):
//
//	"{1}{G}: This creature gets +2/+2 until end of turn. Activate only
//	 once each turn.
//	 Madness {0} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Kraven's Cats' once-a-turn pump at {1}{G} (OncePerTurnActivation, CR
// 602.5b), and madness for a real {0}, which the engine grows from the
// one string (CR 702.35a, #657): a discarded Rootwalla is cast from
// exile for nothing.
//
// No simplification.
func init() {
	const label = "{1}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "b2ec080f-ad5d-4274-91e3-bc6ac89fc3e4",
		Name:         "Basking Rootwalla",
		Completeness: CompletenessFull,
		Madness:      "{0}",
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{1}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Basking Rootwalla — +2/+2 until end of turn"),
		}},
	})
}
