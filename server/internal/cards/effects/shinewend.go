package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shinewend — Creature — Elemental {1}{W}, 0/0 (Morningtide, #5592):
//
//	"Flying
//	 This creature enters with a +1/+1 counter on it.
//	 {1}{W}, Remove a +1/+1 counter from this creature: Destroy target
//	 enchantment."
//
// District Mascot's shape: the entry counter is a self-replacement
// (CR 614.1c), and the counter removal is part of the cost, paid at
// announce (CR 602.2b). Spending its only counter leaves a 0/0, which
// dies to the state check after the ability is activated (CR 704.5f); the
// ability still resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b3c4a6c-44ac-4718-ac1e-9d503523e89d",
		Name:            "Shinewend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 1, "Shinewend: enters with a +1/+1 counter"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Remove a +1/+1 counter from this creature: Destroy target enchantment.",
			Cost:    Plus(ManaCost("{1}{W}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
