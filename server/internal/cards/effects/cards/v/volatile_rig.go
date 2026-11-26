package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volatile Rig — Artifact Creature — Construct {4}, 4/4 (Return to Ravnica,
// S82 OLD):
//
//	"Trample
//	 This creature attacks each combat if able.
//	 Whenever this creature is dealt damage, flip a coin. If you lose the
//	 flip, sacrifice this creature.
//	 When this creature dies, flip a coin. If you lose the flip, it deals 4
//	 damage to each creature and each player."
//
// Both flips are the controller's called coin flips (Goblin Kaboomist's
// keyed continuation). The damage trigger waits for the damage to finish, so
// two blockers hitting it together are one trigger (CR 603.2c; the 2012
// ruling); a lost flip sacrifices the Rig only while it is still the same
// permanent. The dies trigger's damage is dealt by the Rig as it last existed
// (CR 608.2h) to every creature and every player, its controller included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba47a07e-890c-411a-9fd0-947c2f738da2",
		Name:            "Volatile Rig",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4}},
		Static:          []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B35SelfWasDealtDamage(ev, source)
				},
				Key:        "Volatile Rig — flip a coin; if you lose the flip, sacrifice it",
				AtBatchEnd: effects.ThatMuchDamage(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Volatile Rig — call the coin flip",
						ThenKey:  volatileRigDamaged,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				},
			},
			effects.WhenThisDies("Volatile Rig — flip a coin; if you lose the flip, it deals 4 damage to each creature and each player",
				func(g *game.Game, item *game.StackItem) error {
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Volatile Rig — call the coin flip",
						ThenKey:  volatileRigDied,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// volatileRigDamaged is "if you lose the flip, sacrifice this creature". An
// on-disk key: never renamed, never reused.
var volatileRigDamaged = effects.CoinFlipThen("coin-flip/volatile-rig-damaged", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
})

// volatileRigDied is "if you lose the flip, it deals 4 damage to each
// creature and each player". An on-disk key: never renamed, never reused.
var volatileRigDied = effects.CoinFlipThen("coin-flip/volatile-rig-died", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return nil
	}
	return effects.B23DamageEachCreatureAndEachPlayer(ctx, 4)
})
