package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Growth Cycle — Instant {1}{G} (Core Set 2020):
//
//	"Target creature gets +3/+3 until end of turn. It gets an additional +2/+2
//	 until end of turn for each card named Growth Cycle in your graveyard."
//
// Rite of Flame's count, read as the spell resolves and only in the caster's
// graveyard: the spell is still on the stack, so the first copy gives +3/+3
// and does not count itself. The bonus is one pump of the whole size (CR
// 611.2), so a target that left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec714e79-10ec-43ea-a0ff-f8c9083fe015",
		Name:         "Growth Cycle",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			named := effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool { return c.Name == "Growth Cycle" })
			n := 3 + 2*len(named)
			return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Growth Cycle — pump"}.Apply(ctx)
		},
	})
}
