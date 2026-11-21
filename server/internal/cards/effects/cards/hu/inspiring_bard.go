package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Bard — Creature — Elf Bard {3}{G}, 3/3 (AFR):
//
//	"When this creature enters, choose one —
//	 • Bardic Inspiration — Target creature gets +2/+2 until end of turn.
//	 • Song of Rest — You gain 3 life."
//
// A modal ETB (CR 603.3c): the first bullet's target is chosen with the mode.
func init() {
	bard := effects.WhenThisEnters("Inspiring Bard — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bard.Modes = effects.ChooseOne(
		effects.ModeDoing("Bardic Inspiration — Target creature gets +2/+2 until end of turn.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2,
					Label: "Inspiring Bard — Bardic Inspiration: +2/+2 until end of turn"}.Apply(ctx)
			}),
		effects.ModeDoing("Song of Rest — You gain 3 life.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "88443470-f696-4f9b-b34d-50be8f1d4bc2",
		Name:         "Inspiring Bard",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{bard},
	})
}
