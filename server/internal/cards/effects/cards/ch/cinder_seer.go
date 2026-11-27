package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cinder Seer — {3}{R} Creature — Human Wizard 1/1 (Urza's Destiny, S82 OLD batch 704, #5621):
//
//	"{2}{R}, {T}: Reveal any number of red cards in your hand. This creature
//	 deals X damage to any target, where X is the number of cards revealed
//	 this way."
//
// The cards are chosen and revealed as the ability resolves (S82OldB704RevealRedForDamage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64b63797-9db0-41fa-b2eb-fbf51de0e337",
		Name:         "Cinder Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Reveal any number of red cards in your hand. This creature deals X damage to any target, where X is the number of cards revealed this way.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.S82OldB704RevealRedForDamage("Cinder Seer"),
		}},
	})
}
