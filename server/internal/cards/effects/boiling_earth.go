package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boiling Earth — Sorcery {1}{R}:
//
//	"Boiling Earth deals 1 damage to each creature your opponents control.
//	 Awaken 4—{6}{R}"
//
// ADR 0135 §3 (#2411): the damage, then, if the awaken cost was paid, four
// +1/+1 counters on target land you control, which becomes a 0/0
// Elemental creature with haste (CR 702.113a). The spell's only target is
// the awaken land, so a cast for {1}{R} has none (CR 702.113b) and an
// awaken cast whose land is gone does not resolve (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fb34b671-b61a-47e2-90fa-dbe5cf6d3743",
		Name:         "Boiling Earth",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true}},
		AlternativeCosts: []game.AlternativeCost{
			Awaken(4, "{6}{R}", nil),
		},
		OnResolve: AwakenAfter(4, func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), OpponentControls()), 1)
		}),
	})
}
