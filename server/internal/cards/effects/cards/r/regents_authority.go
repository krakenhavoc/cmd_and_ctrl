package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Regent's Authority — Instant {W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Target creature gets +2/+2 until end of turn. If it's an enchantment
//	 creature or legendary creature, instead put a +1/+1 counter on it and
//	 it gets +1/+1 until end of turn."
//
// The condition is read off the target's current types as the spell
// resolves. Either branch is exactly one of the two printed outcomes, never
// both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "54402930-84ea-4c3e-9880-ca631e47485f",
		Name:         "Regent's Authority",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if ok && effects.Or(effects.Enchantment(), effects.Legendary())(ctx.Game, ctx.Controller(), c) {
				if err := (effects.AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1,
					Label: "Regent's Authority — +1/+1"}.Apply(ctx)
			}
			return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2,
				Label: "Regent's Authority — +2/+2"}.Apply(ctx)
		},
	})
}
