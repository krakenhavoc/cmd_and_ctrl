package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clockwork Beetle — Artifact Creature — Insect {1}, 0/0 (Mirrodin,
// #5606):
//
//	"This creature enters with two +1/+1 counters on it.
//	 Whenever this creature attacks or blocks, remove a +1/+1 counter
//	 from it at end of combat."
//
// The entry counters are Oathsworn Knight's replacement; the combat
// clause is the Clockwork trigger (clockwork.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b9ddff4-82f5-4068-8105-116b8a6180d5",
		Name:         "Clockwork Beetle",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Clockwork Beetle: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			clockworkLosesACounterAtEndOfCombat("Clockwork Beetle"),
		},
	})
}
