package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawnbringer Cleric — Creature — Human Cleric {1}{W}, 1/3 (AFR):
//
//	"When this creature enters, choose one —
//	 • Cure Wounds — You gain 2 life.
//	 • Dispel Magic — Destroy target enchantment.
//	 • Gentle Repose — Exile target card from a graveyard."
func init() {
	cleric := effects.WhenThisEnters("Dawnbringer Cleric — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	cleric.Modes = effects.ChooseOne(
		effects.ModeDoing("Cure Wounds — You gain 2 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
		effects.ModeDoing("Dispel Magic — Destroy target enchantment.",
			effects.TargetPermanent("target enchantment", effects.Enchantment()),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("Gentle Repose — Exile target card from a graveyard.",
			effects.TargetCardInGraveyard("target card from a graveyard"),
			effects.MurdocksExileTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "0146f430-da82-48d3-a576-c2b05101797f",
		Name:         "Dawnbringer Cleric",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{cleric},
	})
}
