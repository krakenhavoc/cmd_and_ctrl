package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heaven // Earth — split card with aftermath, oracle
// 03d3fee8-f7de-4152-8c18-6d279cc6c125:
//
//	Heaven {X}{G} Instant: "Heaven deals X damage to each creature with
//	 flying."
//	Earth {X}{R}{R} Sorcery — Aftermath: "Earth deals X damage to each
//	 creature without flying."
//
// Two specs, Earth under "<oracle_id>#1" as Dusk // Dawn is; aftermath is
// the engine's (CR 702.127a). Each half is Magmaquake's mass damage with a
// flying test read through the layered view as the spell resolves. Each
// half has its own X, paid when that half is cast.
//
// No simplification.
func init() {
	const oracle = "03d3fee8-f7de-4152-8c18-6d279cc6c125"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Heaven",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), ctx.X())
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Earth",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.WithoutKeyword("flying")), ctx.X())
		},
	})
}
