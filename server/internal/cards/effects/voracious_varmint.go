package effects

// Voracious Varmint — Creature — Varmint {1}{G}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Vigilance
//	 {1}, Sacrifice this creature: Destroy target artifact or
//	 enchantment."
//
// Cathar Commando's activated ability on a vigilant 2/2: a mana-plus-
// sacrifice-self cost paid at announce, with a real target clause chosen
// then, so the destroy uses the stack and an artifact or enchantment
// gone by resolution leaves it nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd785a7e-df65-48e6-843b-7831f3c1e5d3",
		Name:            "Voracious Varmint",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyChosenPermanent,
		}},
	})
}
