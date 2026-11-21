package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reverent Howl — Instant {2}{B} (The Hobbit, HOB):
//
//	"Choose one —
//	 • Target player draws two cards and loses 2 life.
//	 • Target creature gets +2/+2 and gains lifelink until end of turn."
//
// The first bullet is Sign in Blood; the second is Blossoming Defense's
// pump-and-keyword pair, two layers (7c and 6) so two primitives. Each bullet
// reads its own target group (ModeTarget), and a target that left or became
// illegal by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86214b4c-a3ad-4372-b5fe-8f742a7a3d4c",
		Name:         "Reverent Howl",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target player draws two cards and loses 2 life.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					if err := (effects.DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
						return err
					}
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
				}), effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2})),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +2/+2 and gains lifelink until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					const label = "Reverent Howl — +2/+2 and lifelink until end of turn"
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: label}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"lifelink"}, Label: label}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "lifelink"))),
		),
	})
}
