package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnbringer Cleric — Creature — Human Cleric {1}{W}, 1/3 (AFR):
//
//	"When this creature enters, choose one —
//	 • Cure Wounds — You gain 2 life.
//	 • Dispel Magic — Destroy target enchantment.
//	 • Gentle Repose — Exile target card from a graveyard."
func init() {
	cleric := WhenThisEnters("Dawnbringer Cleric — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	cleric.Modes = ChooseOne(
		ModeDoing("Cure Wounds — You gain 2 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
		ModeDoing("Dispel Magic — Destroy target enchantment.",
			TargetPermanent("target enchantment", Enchantment()),
			DestroyTheModesTarget),
		ModeDoing("Gentle Repose — Exile target card from a graveyard.",
			TargetCardInGraveyard("target card from a graveyard"),
			exileTheModesTarget),
	)
	Register(Spec{
		OracleID:     "0146f430-da82-48d3-a576-c2b05101797f",
		Name:         "Dawnbringer Cleric",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{cleric},
	})
}
