package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a90de2b8-4d71-4d0e-82b3-3b7ab194165b",
		Name:         "Brokers Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature you control gets +1/+0 until end of turn. It deals damage equal to its power to target creature or planeswalker an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Harmful(effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()))),
				),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					mine, ok := effects.ModeClauseTarget(ctx, occ, 0)
					if !ok || mine.Kind != game.TargetCard {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: mine.ID, Power: 1, Label: "Brokers Charm — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					theirs, ok := effects.ModeClauseTarget(ctx, occ, 1)
					if !ok || theirs.Kind != game.TargetCard {
						return nil
					}
					return effects.DealDamage{Source: mine.ID, Target: theirs.ID, Amount: effects.B31CurrentPowerOf(ctx.Game, mine.ID)}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0))),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Draw two cards.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				}), game.Purpose{Draws: 2}),
		),
	})
}
