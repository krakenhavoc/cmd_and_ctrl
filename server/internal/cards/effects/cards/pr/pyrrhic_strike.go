package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "5ea1d89c-8650-4373-9835-e369587020ef",
		Name:          "Pyrrhic Strike",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.OptionalBlight(2)},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Destroy target creature with mana value 3 or greater.",
				effects.TargetCreature("target creature with mana value 3 or greater", effects.ManaValueGE(3)),
				effects.DestroyTheModesTarget),
		).InsteadIf(2, effects.BlightUsed),
	})
}
