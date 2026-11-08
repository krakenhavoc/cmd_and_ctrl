package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Officious Interrogation — Instant {W}{U} (Murders at Karlov Manor,
// #4469):
//
//	"This spell costs {W}{U} more to cast for each target beyond the
//	 first.
//	 Choose any number of target players. Investigate X times, where X is
//	 the total number of creatures those players control."
//
// Call the Coppercoats' shape with any player in place of an opponent:
// the strive-style surcharge (CostsMorePerTargetBeyondFirst) and "any
// number of target players", which may be none, or you. X is counted as
// the spell resolves, over the targets still legal then (CR 608.2b).
// Investigating X times is X Clues (CR 701.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e57e448b-1186-4662-a6b0-01a9ce33b961",
		Name:         "Officious Interrogation",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("any number of target players").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			CostsMorePerTargetBeyondFirst("{W}{U}", "This spell costs {W}{U} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := map[uuid.UUID]bool{}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer {
					players[t.ID] = true
				}
			}
			n := 0
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() && players[c.Controller] {
					n++
				}
			}
			if n == 0 {
				return nil
			}
			return CreateToken{Controller: ctx.Controller(), Template: ClueToken(), N: n}.Apply(ctx)
		},
	})
}
