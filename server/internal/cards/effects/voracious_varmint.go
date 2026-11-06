package effects

// Voracious Varmint — Creature — Varmint {1}{G}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Vigilance
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Cathar Commando's activated ability on a vigilant body. The sacrifice is
// a cost, so the target is destroyed even though the Varmint is gone.
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
