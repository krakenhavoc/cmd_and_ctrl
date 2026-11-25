package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Growth — Instant {1}{G} (Dominaria):
//
//	"Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	 Untap target creature. It gets +2/+2 until end of turn. If this spell was
//	 kicked, that creature gets +4/+4 until end of turn instead."
//
// Written as the printed steps (CR 608.2c): the untap, then one pump whose size
// the kicker sets, so a target that has gone only skips both halves
// (CR 608.2b). The pump is a single +2/+2 or +4/+4, never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "edbe4fdf-9e55-44fa-b3ec-5b1ac615f3a7",
		Name:          "Gift of Growth",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			n := 2
			if ctx.WasKicked() {
				n = 4
			}
			return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Gift of Growth — pump"}.Apply(ctx)
		},
	})
}
