package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breaking Point — Sorcery {1}{R}{R}:
//
//	"Any player may have Breaking Point deal 6 damage to them. If no one
//	 does, destroy all creatures. Creatures destroyed this way can't be
//	 regenerated."
//
// ADR 0150 §3 (#1903): Browbeat's offer with Wrath of God's sweep
// (anyPlayerMayTakeDamageFromThis, wrathDestroyAllCreaturesNoRegen).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a648b9ed-810a-41db-af40-3b5b4658db90",
		Name:         "Breaking Point",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		OnResolve:    anyPlayerMayTakeDamageFromThis("Breaking Point", 6, "the sweep", breakingPointSweeps),
	})
}

// breakingPointSweeps is "If no one does, destroy all creatures", keyed
// (#4264).
var breakingPointSweeps = AnyPlayerConsequences("any-player/breaking-point-sweep", nil,
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		return wrathDestroyAllCreaturesNoRegen(ctx.Item, ctx)
	}, nil)
