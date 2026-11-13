package effects

// Trickster Mage — Creature — Human Spellshaper {U}, 1/1
// (Nemesis, #5618):
//
//	"{U}, {T}, Discard a card: You may tap or untap target artifact, creature, or land."
//
// The spellshaper cost (nemSpellshaperCost) on Tidal Bore's "you may
// tap or untap" prompt (nemMayTapOrUntapTheTarget): the controller
// picks tap it, untap it or leave it as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5a9fd97-535c-4aa4-9777-6cf8a413b6ec",
		Name:         "Trickster Mage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: You may tap or untap target artifact, creature, or land.",
			Cost:    nemSpellshaperCost("{U}"),
			Targets: TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
			Effect:  nemMayTapOrUntapTheTarget("Trickster Mage"),
		}},
	})
}
