package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reluctant Dounguard — Creature — Kithkin Soldier {2}{W}, 4/4 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 Whenever another creature you control enters while this creature has a
//	 -1/-1 counter on it, remove a -1/-1 counter from this creature."
//
// Bristlebane Battler with two counters instead of five, sharing its two
// helpers: "while this creature has a -1/-1 counter on it" is part of the
// trigger event (not an intervening if), and the body removes one counter
// if there is still one. If it enters together with other creatures, its
// trigger fires for each of them, as long as it entered with counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "162dcfa0-36e4-41cd-8322-08e810ff04d4",
		Name:         "Reluctant Dounguard",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Reluctant Dounguard: enters with two -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(AnotherCreatureEnteredUnderYourControl, thisHasAMinusOneCounter),
				"Reluctant Dounguard — remove a -1/-1 counter from this creature",
				removeAMinusOneCounterFromThis),
		},
	})
}
