package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chain Lightning — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "8785f42f-87b9-4a0a-89ce-ea423649ba9c", Name: "Chain Lightning", Completeness: CompletenessFull,
		Targets: TargetAny(), Purpose: ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			t := ts[0]
			tail := withCarry(ctx.ItemTail(chainLightningDamaged), t.ID)
			if ctx.Game.PlayerByIDForEffect(t.ID) != nil {
				tail.Nums = []int{1}
				return ctx.Game.DealDamageToPlayerKeyedForEffect(ctx.Source(), t.ID, 3, tail)
			}
			ref, ok := ctx.Game.PermanentRefForEffect(t.ID)
			if !ok {
				return nil
			}
			tail.Nums = []int{0, ref.Epoch}
			return ctx.Game.DealDamageToCreatureKeyedForEffect(ctx.Source(), t.ID, 3, tail)
		},
	})
}

var chainLightningDamaged = CardTail("then/chain-lightning-offer-red-payment", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.Tail.Nums) == 0 {
		return nil
	}
	player := o.Tail.Carry[0]
	if o.Tail.Nums[0] == 0 {
		if len(o.Tail.Nums) < 2 {
			return nil
		}
		info, ok := ctx.Game.PermanentForEffect(game.ObjectRef{ID: player, Epoch: o.Tail.Nums[1]})
		if !ok {
			return nil
		}
		player = info.Controller
	}
	return MayPay{Chooser: player, Cost: "{R}{R}", Question: "Chain Lightning — pay {R}{R} to be allowed to copy this spell?", Then: chainLightningPaid}.Apply(ctx)
})
var chainLightningPaid = PayUnlessBranches("pay-unless/chain-lightning-offer-copy", func(ctx *Context, r game.PayUnlessAnswer) error {
	return chainOfferCopy(ctx.Game, ctx.Item, r.Chooser, "Chain Lightning")
}, nil)
