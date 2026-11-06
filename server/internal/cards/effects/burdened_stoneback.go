package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burdened Stoneback — Creature — Giant Warrior {1}{W}, 4/4 (Lorwyn
// Eclipsed, #4458):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {1}{W}, Remove a counter from this creature: Target creature gains
//	 indestructible until end of turn. Activate only as a sorcery. (Damage
//	 and effects that say "destroy" don't destroy it. If its toughness is
//	 0 or less, it still dies.)"
//
// Glen Elendra Guardian's shape: the counters are an entry replacement,
// and "a counter" is one counter of a kind the activator names, so once
// the -1/-1 counters are gone a +1/+1 counter it picked up pays too. The
// grant is Old Hob's body. The Stoneback may target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a5e17a0-6f3c-49ed-b328-ed90a46ed6c5",
		Name:         "Burdened Stoneback",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 2, "Burdened Stoneback: enters with two -1/-1 counters"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{W}, Remove a counter from this creature: Target creature gains indestructible until end of turn. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{W}"), RemoveCountersFromThis("", 1)),
			Targets:      TargetCreature("target creature"),
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible")),
			SorcerySpeed: true,
			Effect:       ebTargetCreatureGainsUntilEOT("indestructible", "Burdened Stoneback — indestructible"),
		}},
	})
}
