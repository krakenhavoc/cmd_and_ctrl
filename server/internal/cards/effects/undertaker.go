package effects

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
	Register(Spec{
		OracleID:     "a8ffbf6d-44e9-44e4-846c-4f894078130e",
		Name:         "Undertaker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}, Discard a card: Return target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{B}"), TapCost(), DiscardACard()),
			Targets: targetCreatureInYourGraveyard(),
			Effect:  b27ReturnChosenGraveyardCardToHand,
		}},
	})
}
