package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	creatureOrWalker := Or(Creature(), Planeswalker())
	Register(Spec{
		OracleID:     "2a29d028-d1a4-413f-9c01-d6842294420a",
		Name:         "Fulminous Forte",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Fulminous Forte deals 1 damage to each creature and planeswalker your opponents control.",
				nil, func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, And(creatureOrWalker, OpponentControls()), 1)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}}),
			ModeWithPurpose(ModeDoing("Fulminous Forte deals 5 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", creatureOrWalker),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 5}.Apply(ctx)
				}), ForTargets(DamageToTarget(0, 5))),
		),
	})
}
