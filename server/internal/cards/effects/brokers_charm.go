package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brokers Charm — Instant {G}{W}{U} (Streets of New Capenna):
//
//	"Choose one —
//	 • Target creature you control gets +1/+0 until end of turn. It deals
//	   damage equal to its power to target creature or planeswalker an
//	   opponent controls.
//	 • Destroy target enchantment.
//	 • Draw two cards."
//
// The first bullet is Assert Perfection's bite with the second target
// mandatory: the +1/+0 lands first, the damage reads the layered power
// after it, and your creature is the source. Each target keeps its own
// clause (slot 0, slot 1). If your creature has left neither half
// happens; a recipient that left takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a90de2b8-4d71-4d0e-82b3-3b7ab194165b",
		Name:         "Brokers Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Target creature you control gets +1/+0 until end of turn. It deals damage equal to its power to target creature or planeswalker an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					Harmful(TargetPermanent("target creature or planeswalker an opponent controls", And(Or(Creature(), Planeswalker()), OpponentControls()))),
				),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					mine, ok := ModeClauseTarget(ctx, occ, 0)
					if !ok || mine.Kind != game.TargetCard {
						return nil
					}
					if err := (BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Brokers Charm — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					theirs, ok := ModeClauseTarget(ctx, occ, 1)
					if !ok || theirs.Kind != game.TargetCard {
						return nil
					}
					return DealDamage{Source: mine.ID, Target: theirs.ID, Amount: b31CurrentPowerOf(ctx.Game, mine.ID)}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 1, 0))),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
			ModeWithPurpose(ModeDoing("Draw two cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
		),
	})
}
