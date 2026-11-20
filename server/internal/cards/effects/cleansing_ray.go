package effects

// Cleansing Ray — Sorcery {1}{W} (Rivals of Ixalan, #5531):
//
//	"Choose one —
//	 • Destroy target Vampire.
//	 • Destroy target enchantment."
//
// Each bullet's target clause is its own, so the picker offers Vampires
// or enchantments by mode; the body is the shared DestroyTheModesTarget.
// "Vampire" is any permanent of that type, not only a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50f1948c-fa27-458d-be14-cb0e09e39f63",
		Name:         "Cleansing Ray",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target Vampire.", TargetPermanent("target Vampire", Subtype("Vampire")), DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()), DestroyTheModesTarget),
		),
	})
}
