package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidetsugu's Second Rite — Instant {3}{R} (Foundations, #4465):
//
//	"If target player has exactly 10 life, Hidetsugu's Second Rite deals
//	 10 damage to that player."
//
// Having exactly 10 life is no targeting restriction (the ruling): any
// player is a legal target, and the life total is read as the spell
// resolves. A player at any other total makes it do nothing and finish
// resolving. A target that is no longer legal at resolution is skipped
// like any other (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "749683c9-8cc0-4e8e-8d90-adda5e5f4e93",
		Name:         "Hidetsugu's Second Rite",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 10))},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target := TargetedPlayer(ctx)
			p := ctx.Game.PlayerByIDForEffect(target)
			if p == nil || p.Life != 10 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: target, Amount: 10}.Apply(ctx)
		},
	})
}
