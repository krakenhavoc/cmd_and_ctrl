package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oppressive Will — Instant {2}{U} (Saviors of Kamigawa, #5601):
//
//	"Counter target spell unless its controller pays {1} for each card in
//	 your hand."
//
// Miscalculation's soft counter with a price counted as the spell resolves,
// from the caster's hand, so Oppressive Will itself is not in it
// (CR 608.2h). With an empty hand the price is nothing and the spell is not
// countered. The spell's controller is asked (CR 118.12).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "983a2a50-3634-4783-8576-30916c3d0818",
		Name:         "Oppressive Will",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			cost := effects.GenericMana(effects.SokB1CardsInYourHand(ctx))
			if cost == "" {
				return nil
			}
			return effects.CounterUnlessPaid{
				StackID:  t.ID,
				Cost:     cost,
				Question: "Oppressive Will — pay " + cost + " or your spell is countered",
			}.Apply(ctx)
		},
	})
}
