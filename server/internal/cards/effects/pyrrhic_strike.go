package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pyrrhic Strike — {2}{W} Instant:
//
//	"As an additional cost to cast this spell, you may blight 2. (You
//	 may put two -1/-1 counters on a creature you control.)
//	 Choose one. If this spell's additional cost was paid, choose both
//	 instead.
//	 • Destroy target artifact or enchantment.
//	 • Destroy target creature with mana value 3 or greater."
//
// #1703: OptionalBlight(2) with InsteadIf(2, BlightUsed). The blight is
// paid at CR 601.2h onto a creature the caster names, which may die of
// it and still pays. No simplification.
func init() {
	Register(Spec{
		OracleID:      "5ea1d89c-8650-4373-9835-e369587020ef",
		Name:          "Pyrrhic Strike",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{OptionalBlight(2)},
		Modes: ChooseOne(
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				DestroyTheModesTarget),
			ModeDoing("Destroy target creature with mana value 3 or greater.",
				TargetCreature("target creature with mana value 3 or greater", ManaValueGE(3)),
				DestroyTheModesTarget),
		).InsteadIf(2, BlightUsed),
	})
}
