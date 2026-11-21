package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swarm Shambler — Creature — Fungus Beast {G}, 0/0 (Jumpstart 2022, #5948):
//
//	"This creature enters with a +1/+1 counter on it.
//	 Whenever a creature you control with a +1/+1 counter on it becomes the
//	 target of a spell an opponent controls, create a 1/1 green Insect
//	 creature token.
//	 {1}, {T}: Put a +1/+1 counter on this creature."
//
// The counter is a replacement on its entry (CR 614.1c). The trigger watches
// EventBecomesTarget, which fires once per target slot (CR 115.3): only the
// first event of one spell for one creature counts, only for a SPELL (not an
// ability) of an opponent's, and only for a creature that has a +1/+1 counter
// as it becomes the target (the Shambler included).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f46c72b8-e2e8-42e6-a2e7-ffa5810a9bf3",
		Name:         "Swarm Shambler",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Swarm Shambler: enters with a +1/+1 counter"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, gapB153CreatureWithCounterTargetedBySpell,
				"Swarm Shambler — create a 1/1 green Insect creature token",
				Do(CreateToken{Template: TokenCard("1/1 green Insect"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  plusOneCountersOnThis(1),
		}},
	})
}
