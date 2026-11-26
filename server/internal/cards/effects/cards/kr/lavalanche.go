package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lavalanche — Sorcery {X}{B}{R}{G}:
//
//	"Lavalanche deals X damage to target player or planeswalker and each
//	 creature that player or that planeswalker's controller controls."
//
// X rides the cast (ctx.X()). The player is the target, or the controller of
// the targeted planeswalker, read as the spell resolves and before any damage
// lands (a planeswalker the damage kills is no longer there to say who
// controlled it). The creatures are snapshotted first and every point is one
// simultaneous damage instance (CR 608.2). With its one target illegal the
// spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4475e9a7-4e94-4965-8e7e-211d40a98217",
		Name:         "Lavalanche",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageXToTarget(0)),
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB437XDamageToPlayerOrWalkerAndTheirCreatures(ctx)
		},
	})
}
