package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspiring Bard — Creature — Elf Bard {3}{G}, 3/3 (AFR):
//
//	"When this creature enters, choose one —
//	 • Bardic Inspiration — Target creature gets +2/+2 until end of turn.
//	 • Song of Rest — You gain 3 life."
//
// A modal ETB (CR 603.3c): the first bullet's target is chosen with the mode.
func init() {
	bard := WhenThisEnters("Inspiring Bard — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	bard.Modes = ChooseOne(
		ModeDoing("Bardic Inspiration — Target creature gets +2/+2 until end of turn.",
			TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2,
					Label: "Inspiring Bard — Bardic Inspiration: +2/+2 until end of turn"}.Apply(ctx)
			}),
		ModeDoing("Song of Rest — You gain 3 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "88443470-f696-4f9b-b34d-50be8f1d4bc2",
		Name:         "Inspiring Bard",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{bard},
	})
}
