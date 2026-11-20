package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deserted Temple — Land (EDHREC rank 1603):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Untap target land."
//
// The Cabal Coffers enabler. The untap is an ordinary CR 602
// activated ability with a mana-and-tap cost and a battlefield land
// target — any land, anyone's, as printed — and it uses the stack, so
// it can be responded to. The mana ability is the plain colourless
// tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f12bf9a-6e1a-4377-b4af-e8cabd3ee58a",
		Name:         "Deserted Temple",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Untap target land.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.UntapEachLegalTarget,
		}},
	})
}
