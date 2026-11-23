package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "26b441b9-fdfb-43e2-bc02-b358653eadf4",
		Name:         "Over the Edge",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one", 1, 1,
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())), effects.DestroyTheModesTarget),
			effects.ModeSteps("Target creature you control explores, then it explores again.",
				effects.TargetPermanent("target creature you control", effects.Creature(), effects.YouControl()),
				effects.Instruction(effects.ExploreTheModeTarget), effects.Instruction(effects.ExploreTheModeTarget)),
		),
	})
}
