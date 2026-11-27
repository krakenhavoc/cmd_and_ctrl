package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Bangchuckers — Creature — Goblin Warrior {2}{R}{R}, 2/2 (Magic
// 2012, S82 OLD batch 452, #5573):
//
//	"{T}: Flip a coin. If you win the flip, this creature deals 2 damage
//	 to any target. If you lose the flip, this creature deals 2 damage to
//	 itself."
//
// The target is chosen as the ability goes on the stack, the coin is
// called as it resolves, and the keyed continuation (Goblin Kaboomist's
// shape) deals the damage: to the target on a win, if it is still legal,
// to this creature on a loss, if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2561cc9-3840-4bd7-ad8b-a85d35e47f71",
		Name:         "Goblin Bangchuckers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Flip a coin. If you win the flip, this creature deals 2 damage to any target. If you lose the flip, this creature deals 2 damage to itself.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Goblin Bangchuckers — call the coin flip",
					ThenKey:  goblinBangchuckersFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// goblinBangchuckersFlipped is the flip's result. An on-disk key: never
// renamed, never reused.
var goblinBangchuckersFlipped = effects.CoinFlipThen("coin-flip/s82-old-b452-goblin-bangchuckers", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return effects.DealDamageToTheTarget(2)(ctx.Game, ctx.Item)
	}
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: ctx.Source(), Amount: 2}.Apply(ctx)
})
