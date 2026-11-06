package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Train — Artifact — Vehicle {3}{B}, 4/4 (Final Fantasy, #4462):
//
//	"Trample
//	 Sacrifice another artifact or creature: Put a +1/+1 counter on this
//	 Vehicle. It becomes a Spirit artifact creature in addition to its
//	 other types until end of turn."
//
// The sacrifice is the cost (another permanent, never the Train). The
// counter and the animation are crew's BecomeCreatureUntilEOT with the
// Spirit creature type added, pinned to this object (CR 400.7): a Train
// that left and came back in response gets neither. The Train has no
// crew ability; this is the only way it becomes a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a0c4eb78-bee0-42c6-ab1e-64bacde87960",
		Name:            "Phantom Train",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another artifact or creature: Put a +1/+1 counter on this Vehicle. It becomes a Spirit artifact creature in addition to its other types until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerAnimate | game.AnswerPump},
			Cost:    SacrificeAnotherN(1, "another artifact or creature", Or(Artifact(), Creature())),
			Effect: Do(
				CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				BecomeCreatureUntilEOT{Subtypes: []string{"Spirit"}, Label: "Phantom Train — a Spirit artifact creature"},
			),
		}},
	})
}
