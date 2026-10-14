package effects

// Destroy Evil — Instant {1}{W}:
//
//	"Choose one —
//	 • Destroy target creature with toughness 4 or greater.
//	 • Destroy target enchantment."
//
// Two destroy bullets on the shared DestroyTheModesTarget body. The
// toughness is judged at announce and again at resolution (CR 608.2b),
// so a creature shrunk below 4 in response is no longer a legal target
// and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b75e667a-1ceb-445a-acf4-ae1fd2dae9e7",
		Name:         "Destroy Evil",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target creature with toughness 4 or greater.",
				TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
				DestroyTheModesTarget),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
		),
	})
}
