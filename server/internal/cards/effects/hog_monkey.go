package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hog-Monkey — Creature — Boar Monkey {2}{B}, 3/2:
//
//	"At the beginning of combat on your turn, target creature you
//	 control with a +1/+1 counter on it gains menace until end of turn.
//	 Exhaust — {5}: Put two +1/+1 counters on this creature. (Activate
//	 each exhaust ability only once.)"
//
// The target must carry a +1/+1 counter as the trigger goes on the stack
// (CR 603.3d, the ruling): with none on the battlefield the trigger is
// removed, and a counter put on in response comes too late. It is
// rechecked as it resolves (CR 608.2b). The exhaust is the engine's
// once-per-object marker (CR 702.177a): a Hog-Monkey that left and came
// back is a new object and may activate it again (the ruling).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	const label = "Hog-Monkey — target creature you control with a +1/+1 counter on it gains menace until end of turn"
	Register(Spec{
		OracleID:     "e11591ee-7074-488e-8d2f-f8c89e385a55",
		Name:         "Hog-Monkey",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(label, firstTargetGainsUntilEndOfTurn(0, label, "menace")),
				TargetCreature("target creature you control with a +1/+1 counter on it",
					YouControl(), CreatureWithCounter(game.CounterPlusOne))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {5}: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    ManaCost("{5}"),
			Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2}),
		}},
	})
}
