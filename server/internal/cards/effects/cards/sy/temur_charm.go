package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temur Charm — Instant {G}{U}{R} (Commander 2020, S82 OLD, #5513):
//
//	"Choose one —
//	 • Target creature you control gets +1/+1 until end of turn. It fights
//	   target creature you don't control.
//	 • Counter target spell unless its controller pays {3}.
//	 • Creatures with power 3 or less can't block this turn."
//
// The first bullet is Inscription of Abundance's fight with the pump added
// before it (the +1/+1 comes first, so it counts in the fight); it does
// nothing to the fight if either creature is gone (CR 608.2b, 701.14b). The
// second is Mystic Confluence's counter-unless-paid. The third is a live
// "can't block" over every creature with power 3 or less
// (game.ScopeCreaturesPowerThreeOrLess): the set is not locked (CR 611.2c),
// so a creature that arrives later, or whose power later falls to 3, can't
// block either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "418bc6c8-972b-4f07-aac0-8d003b9c1b17",
		Name:         "Temur Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature you control gets +1/+1 until end of turn. It fights target creature you don't control.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Slot == 0 && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Temur Charm — +1/+1"}).Apply(ctx); err != nil {
								return err
							}
						}
					}
					return effects.FightTheModesTargets(item, ctx, occ)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
			effects.ModeDoing("Counter target spell unless its controller pays {3}.",
				effects.TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{3}",
						Question: "Temur Charm — pay {3} or your spell is countered",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Creatures with power 3 or less can't block this turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.RestrictUntilEOT{Scope: game.ScopeCreaturesPowerThreeOrLess, Restrictions: game.CantBlock,
						Label: "Temur Charm — creatures with power 3 or less can't block"}.Apply(ctx)
				}),
		),
	})
}
