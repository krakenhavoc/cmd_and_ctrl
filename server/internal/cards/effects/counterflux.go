package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Counterflux — Instant {U}{U}{R}:
//
//	"This spell can't be countered.
//	 Counter target spell you don't control.
//	 Overload {1}{U}{U}{R} (You may cast this spell for its overload
//	 cost. If you do, change "target" in its text to "each.")"
//
// Three printed clauses, three existing shapes. CantBeCountered is
// the S23 rider — a counterspell aimed at Counterflux resolves and
// does nothing (server/internal/game/cant_be_countered.go). "You
// don't control" is OpponentControls(), the same predicate every
// other "target X you don't control" clause in the catalog composes
// with; a spell's Card.Controller is stamped from its owner when it
// enters a hand and carries unchanged onto the stack for an ordinary
// cast, so this reads correctly without any spell-specific plumbing.
//
// Overload is the S22 alternative-cost constructor: it bundles the
// {1}{U}{U}{R} price with deleting the target clause (CR 702.96), so
// an overloaded cast has no target to fizzle and nothing for
// hexproof to stop. The swept set is every spell an opponent
// controls — CounterAllMatching snapshots the stack once, so a
// counter early in the sweep can't shrink the set a later one reads.
// Counterflux itself is never swept: it's a spell the CASTER
// controls, and OpponentControls() already excludes that.
func init() {
	Register(Spec{
		OracleID:        "8c983da8-1436-4c06-af9c-91b72cab48c1",
		Name:            "Counterflux",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetSpell("target spell you don't control", OpponentControls()),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{1}{U}{U}{R}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return CounterAllMatching{Match: OpponentControls()}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
