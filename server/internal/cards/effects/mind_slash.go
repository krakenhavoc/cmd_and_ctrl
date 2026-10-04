package effects

// Mind Slash — Enchantment {1}{B}{B}:
//
//	"{B}, Sacrifice a creature: Target opponent reveals their hand. You
//	 choose a card from it. That player discards that card. Activate
//	 only as a sorcery."
//
// Corpse Traders' ability on an enchantment for {B}: the revealed-hand
// pick (ADR 0116) with no filter, chosen by the activator (CR 113.8),
// at sorcery timing (CR 602.5d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47e79913-322e-4ba6-9bf8-3b4c1c82a6c4",
		Name:         "Mind Slash",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{B}, Sacrifice a creature: Target opponent reveals their hand. You choose a card from it. That player discards that card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{B}"), SacrificeACreature()),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect:       TargetRevealsYouChooseDiscardAbility(nil, "card"),
		}},
	})
}
