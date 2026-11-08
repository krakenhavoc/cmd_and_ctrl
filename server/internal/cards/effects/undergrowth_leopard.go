package effects

// Undergrowth Leopard — Creature — Cat {1}{G}, 2/2 (Tarkir: Dragonstorm,
// #4463):
//
//	"Vigilance
//	 {1}, Sacrifice this creature: Destroy target artifact or enchantment."
//
// Vigilance is a printed keyword; the activated ability is Caustic
// Caterpillar's with a {1} cost. The Leopard is sacrificed as the ability
// is activated, and the target is checked again as it resolves; an
// indestructible pick survives, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "511fa8a3-0a65-47fc-9f3b-57eaef2698cd",
		Name:            "Undergrowth Leopard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Destroy target artifact or enchantment.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
