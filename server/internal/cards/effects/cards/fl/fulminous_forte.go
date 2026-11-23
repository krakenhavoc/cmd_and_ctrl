package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fulminous Forte — Instant {2}{R}:
//
//	"Choose one —
//	 • Fulminous Forte deals 1 damage to each creature and planeswalker
//	   your opponents control.
//	 • Fulminous Forte deals 5 damage to target creature or planeswalker."
//
// The first bullet is a sweep that names no target, so it takes
// nothing from the board a player protects with hexproof; the second
// is a plain targeted burn.
//
// No simplifications.
func init() {
	creatureOrWalker := effects.Or(effects.Creature(), effects.Planeswalker())
	effects.Register(effects.Spec{
		OracleID:     "2a29d028-d1a4-413f-9c01-d6842294420a",
		Name:         "Fulminous Forte",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Fulminous Forte deals 1 damage to each creature and planeswalker your opponents control.",
				nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.And(creatureOrWalker, effects.OpponentControls()), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}}),
			effects.ModeWithPurpose(effects.ModeDoing("Fulminous Forte deals 5 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", creatureOrWalker),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}.Apply(ctx)
				}), effects.ForTargets(effects.DamageToTarget(0, 5))),
		),
	})
}
