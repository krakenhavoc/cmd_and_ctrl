package effects

// Seismic Mage — Creature — Human Spellshaper {3}{R}, 1/1
// (Mercadian Masques, #5619):
//
//	"{2}{R}, {T}, Discard a card: Destroy target land."
//
// The spellshaper cost (Kris Mage's: mana, {T} and DiscardACard, the
// card discarded as the ability is activated, CR 602.2b) on the shared
// destroyFirstLegalTarget body: a land that has left the battlefield by
// resolution is not destroyed (CR 608.2b). The {T} is a creature's, so
// the Mage waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1f7e43a-f126-47d8-bd59-339ae052e81b",
		Name:         "Seismic Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: Destroy target land.",
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), DiscardACard()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
