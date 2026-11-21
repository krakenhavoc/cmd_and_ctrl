package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quandrix Charm — Instant {G}{U} (Secrets of Strixhaven, #4456):
//
//	"Choose one —
//	 • Counter target spell unless its controller pays {2}.
//	 • Destroy target enchantment.
//	 • Target creature has base power and toughness 5/5 until end of turn."
//
// The counter bullet is the pay-or-lose prompt to the spell's controller
// (CounterUnlessPaid); the destroy bullet is the shared modal destroy; the
// last sets the creature's BASE power and toughness (layer 7b, CR 613.4b)
// until the turn ends (CR 514.2), so counters and pumps still apply on top.
// Each bullet is checked against its own target at announce and at
// resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f43e36ea-6bf5-45bd-a4ad-cc507ae45641",
		Name:         "Quandrix Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Counter target spell unless its controller pays {2}.",
				effects.TargetSpell("target spell"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.CounterUnlessPaid{
						StackID:  t.ID,
						Cost:     "{2}",
						Question: "Quandrix Charm — pay {2} or your spell is countered?",
					}.Apply(ctx)
				}),
			effects.ModeDoing("Destroy target enchantment.",
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
				effects.DestroyTheModesTarget),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature has base power and toughness 5/5 until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, t.ID, nil,
						"Quandrix Charm — base power and toughness 5/5",
						game.SetBasePowerMod(5), game.SetBaseToughnessMod(5))
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		),
	})
}
