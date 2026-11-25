package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hapatra's Mark — {G} Instant (Amonkhet, S82 old batch 247):
//
//	"Target creature you control gains hexproof until end of turn. Remove all
//	 -1/-1 counters from it."
//
// Both halves apply to the one target as the spell resolves; a target that is
// no longer legal gets neither (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "143d3fcf-2152-4911-b95a-d23b91662f33",
		Name:         "Hapatra's Mark",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Hapatra's Mark — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || c.Counters[game.CounterMinusOne] <= 0 {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterMinusOne, N: -c.Counters[game.CounterMinusOne]}.Apply(ctx)
		},
	})
}
