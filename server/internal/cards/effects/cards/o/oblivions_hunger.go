package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oblivion's Hunger — Instant {1}{B} (Zendikar Rising, S82 OLD #5508):
//
//	"Target creature you control gains indestructible until end of turn.
//	 Draw a card if that creature has a +1/+1 counter on it."
//
// The counter is looked for as the spell resolves (CR 608.2c), after the
// creature has gained indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "517be4da-9aa0-4a83-a559-962df0450f2c",
		Name:         "Oblivion's Hunger",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Oblivion's Hunger — indestructible"}).Apply(ctx); err != nil {
				return err
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found || c.Counters[game.CounterPlusOne] == 0 {
				return nil
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
