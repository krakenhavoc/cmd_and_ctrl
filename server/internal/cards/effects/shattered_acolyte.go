package effects

// Shattered Acolyte — Creature — Dwarf Warlock {1}{W}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Lifelink
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Cathar Commando's ability (destroyChosenPermanent): the sacrifice is a cost,
// so the destroy still happens after the Acolyte is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57f58d52-0014-4cac-94fc-668063cc147a",
		Name:            "Shattered Acolyte",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyChosenPermanent,
		}},
	})
}
