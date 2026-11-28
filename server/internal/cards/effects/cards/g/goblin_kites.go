package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Kites — Enchantment {1}{R} (Fallen Empires, S82 OLD batch 852, #5637):
//
//	"{R}: Target creature you control with toughness 2 or less gains flying until
//	 end of turn. Flip a coin at the beginning of the next end step. If you lose the
//	 flip, sacrifice that creature."
//
// The grant is Abandon Reason's. The flip is a delayed trigger at the next end
// step (CR 603.7) that carries the creature as the object it is now (its id and
// CR 400.7 epoch, as Incandescent Soulstoke's sacrifice does); its controller calls
// the coin (Fickle Efreet's keyed flip), and a lost flip sacrifices that creature
// if it is still that object and still under that controller's control. A won
// flip, or a creature that has gone, does nothing further; the flip is made
// regardless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6efd11c8-be8a-4d44-a776-66d058638ad1",
		Name:         "Goblin Kites",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature you control with toughness 2 or less gains flying until end of turn. Flip a coin at the beginning of the next end step. If you lose the flip, sacrifice that creature.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature you control with toughness 2 or less", effects.YouControl(), effects.ToughnessLE(2)),
			Effect:  goblinKitesResolve,
		}},
	})
}

// goblinKitesResolve grants flying to the target still legal and schedules the
// flip at the next end step, naming the creature as the object it is now.
func goblinKitesResolve(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Goblin Kites — flying"}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      "Goblin Kites — flip a coin; if you lose the flip, sacrifice that creature",
		Body:       goblinKitesFlipBody,
		Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
}

// goblinKitesFlipBody is the delayed trigger: its controller flips a coin, the
// creature and its epoch riding the call. An on-disk key: never renamed, never reused.
var goblinKitesFlipBody = game.DelayedBody("s82-old-b852/goblin-kites-flip", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Goblin Kites — call the coin flip (lose it and the creature is sacrificed)",
		ThenKey:  goblinKitesFlipped,
		Carry:    []uuid.UUID{p.Object.ID},
		Nums:     []int{p.Object.Epoch},
		Asked:    game.NewAskedItem(item),
	})
	return nil
})

// goblinKitesFlipped is the settled call: a lost flip sacrifices the creature if it
// is still the object that was granted flying and is still under the controller's
// control. An on-disk key: never renamed, never reused.
var goblinKitesFlipped = effects.CoinFlipThen("coin-flip/s82-old-b852-goblin-kites", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) || len(r.Carry) == 0 || len(r.Nums) == 0 {
		return nil
	}
	id := r.Carry[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !effects.OnBattlefield(ctx.Game, id) || c.ObjectEpoch != r.Nums[0] || c.Controller != ctx.Controller() {
		return nil
	}
	return effects.SacrificePermanent{Target: id}.Apply(ctx)
})
