package effects

// Over the Edge — Sorcery {1}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"Choose one —
//	 • Destroy target artifact or enchantment.
//	 • Target creature you control explores, then it explores again."
//
// The second explore is its own step, so it waits for the first's
// graveyard question (CR 608.2c). Each bullet's target is its own group
// (CR 601.2c), re-checked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26b441b9-fdfb-43e2-bc02-b358653eadf4",
		Name:         "Over the Edge",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one", 1, 1,
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())), DestroyTheModesTarget),
			ModeSteps("Target creature you control explores, then it explores again.",
				TargetPermanent("target creature you control", Creature(), YouControl()),
				Instruction(exploreTheModeTarget), Instruction(exploreTheModeTarget)),
		),
	})
}
