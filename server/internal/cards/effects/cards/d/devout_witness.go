package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devout Witness — Creature — Human Spellshaper {2}{W}, 2/2
// (Mercadian Masques, #5619):
//
//	"{1}{W}, {T}, Discard a card: Destroy target artifact or
//	 enchantment."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on Caustic
// Caterpillar's destroy (destroyFirstLegalTarget): the target is
// re-checked as it resolves (CR 608.2b), and an indestructible one
// survives. The {T} is a creature's, so the Witness waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83647f5b-2b0f-4f9d-83c5-75c52460f35b",
		Name:         "Devout Witness",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}, Discard a card: Destroy target artifact or enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
