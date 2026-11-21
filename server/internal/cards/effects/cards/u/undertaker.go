package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undertaker — Creature — Human Spellshaper {1}{B}, 1/1
// (Mercadian Masques, #5619):
//
//	"{B}, {T}, Discard a card: Return target creature card from your
//	 graveyard to your hand."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on Memorial to
// Folly's graveyard return. The target is chosen before the cost is
// paid (CR 601.2c, 601.2h), so the card discarded for it can't be the
// card it returns. A target that has left the graveyard by resolution
// is illegal and nothing happens (CR 608.2b). The {T} is a creature's,
// so the Undertaker waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8ffbf6d-44e9-44e4-846c-4f894078130e",
		Name:         "Undertaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Discard a card: Return target creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreatureInYourGraveyard(),
			Effect:  effects.B27ReturnChosenGraveyardCardToHand,
		}},
	})
}
