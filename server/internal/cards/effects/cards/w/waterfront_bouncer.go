package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waterfront Bouncer — Creature — Merfolk Spellshaper {1}{U}, 1/1
// (Mercadian Masques, #5619):
//
//	"{U}, {T}, Discard a card: Return target creature to its owner's
//	 hand."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on the shared
// bounceChosenTarget body: the creature goes to its OWNER's hand, and
// one that has left the battlefield by resolution is not followed (CR
// 608.2b). The {T} is a creature's, so the Bouncer waits out summoning
// sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b53a8cde-494c-43d0-8d08-7bdfe2e84064",
		Name:         "Waterfront Bouncer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Return target creature to its owner's hand.",
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BounceChosenTarget,
		}},
	})
}
