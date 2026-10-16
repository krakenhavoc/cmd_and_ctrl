package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loch Mare — Creature — Horse Serpent {1}{U}, 4/5 (#3987):
//
//	"This creature enters with three -1/-1 counters on it.
//	 {1}{U}, Remove a counter from this creature: Draw a card.
//	 {2}{U}, Remove two counters from this creature: Tap target
//	 creature. Put a stun counter on it."
//
// Neither cost prints a kind. "A counter" is one counter of a kind the
// activator names (RemoveCountersFromThis("", 1)); "two counters" may be
// two of one kind or one each of two (RemoveCountersOfAnyKindsFromThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "631eb3e9-a3e0-4341-ad2d-e79142ce537d",
		Name:         "Loch Mare",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 3, "Loch Mare: enters with three -1/-1 counters"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{U}, Remove a counter from this creature: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{1}{U}"), RemoveCountersFromThis("", 1)),
				Effect:  Do(DrawCards{N: 1}),
			},
			{
				Label:   "{2}{U}, Remove two counters from this creature: Tap target creature. Put a stun counter on it.",
				Cost:    Plus(ManaCost("{2}{U}"), RemoveCountersOfAnyKindsFromThis(2)),
				Targets: TargetCreature("target creature"),
				Effect:  tapTargetAndStunIt,
			},
		},
	})
}
