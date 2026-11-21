package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kris Mage — Creature — Human Spellshaper {R}, 1/1
// (Mercadian Masques, #5619):
//
//	"{R}, {T}, Discard a card: This creature deals 1 damage to any
//	 target."
//
// The spellshaper cost (Charm Peddler's: mana, {T} and DiscardACard,
// the card discarded as the ability is activated, CR 602.2b) on the
// shared sourceDealsDamageToEachLegalTarget body at 1. The {T} is a
// creature's, so the Mage waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d53623b0-f7d8-401f-a8f1-80034f6f20df",
		Name:         "Kris Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Discard a card: This creature deals 1 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
