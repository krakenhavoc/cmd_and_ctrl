package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleanup Crew — Creature — Human Citizen {4}{G}{G}, 4/4 (Streets of New
// Capenna):
//
//	"When this creature enters, choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Exile target card from a graveyard.
//	 • You gain 4 life."
//
// A modal enters trigger (Hivespine Wolverine's shape): the mode and its
// target are chosen as the trigger is put on the stack (CR 700.2b), and a
// bullet whose target has gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Cleanup Crew — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Destroy target artifact.",
			effects.TargetPermanent("target artifact", effects.Artifact()),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("Destroy target enchantment.",
			effects.TargetPermanent("target enchantment", effects.Enchantment()),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("Exile target card from a graveyard.",
			effects.TargetCardInGraveyard("target card from a graveyard"),
			effects.ExileTheModesTarget),
		effects.ModeDoing("You gain 4 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "12e0321b-1cd7-4519-8929-3261cf11a1a9",
		Name:         "Cleanup Crew",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
