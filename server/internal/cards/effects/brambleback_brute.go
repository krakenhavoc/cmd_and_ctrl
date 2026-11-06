package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brambleback Brute — Creature — Giant Warrior {2}{R}, 4/5 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {1}{R}, Remove a counter from this creature: Target creature can't
//	 block this turn. Activate only as a sorcery."
//
// Glen Elendra Guardian's shape: the counters are an entry replacement,
// and "a counter" is one counter of a kind the activator names, so once
// the -1/-1 counters are gone a +1/+1 counter it picked up pays too. The
// restriction is Ahn-Crop Crasher's: the target can't be declared as a
// blocker until the turn ends. With the target gone by resolution the
// ability does nothing.
//
// No simplification.
func init() {
	const label = "Brambleback Brute — target creature can't block this turn"
	Register(Spec{
		OracleID:     "eb25920a-3162-4838-b092-09cbd43332fd",
		Name:         "Brambleback Brute",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Brambleback Brute: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{R}, Remove a counter from this creature: Target creature can't block this turn. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{R}"), RemoveCountersFromThis("", 1)),
			Targets:      TargetCreature("target creature"),
			Purpose:      ForTargets(HarmUntilEOT(0, 0, 0)),
			SorcerySpeed: true,
			Effect:       targetCreatureCantBlockThisTurn(label),
		}},
	})
}
