package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// River's Rebuke — Sorcery {4}{U}{U}:
//
//	"Return all nonland permanents target player controls to their
//	 owner's hand."
//
// The one-player wipe. Six mana to remove a single opponent's entire
// board — creatures, rocks, enchantments, planeswalkers — which at a
// four-player table is the political card: it answers the player who
// is winning without touching the two who aren't, and they notice.
//
// # The predicate is built from the target, not from the caster
//
// This is why ControlledBy exists alongside YouControl and
// OpponentControls. Those two are relative to the CASTER; the sweep
// here is relative to whoever was targeted, who may be any player
// including yourself. The predicate is therefore constructed at
// resolution, closing over the resolved target's ID.
//
// The target is re-checked before OnResolve runs (CR 608.2b), and
// the "return all" is read against control AS THE SPELL RESOLVES —
// a permanent donated away in response is no longer theirs and
// survives.
func init() {
	Register(Spec{
		OracleID: "c52cfb41-18f3-4e73-b5e7-d75baf74e578",
		Name:     "River's Rebuke",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			victim := item.Targets[0].ID
			return BounceAllMatching{
				Match: And(Nonland(), ControlledBy(victim)),
			}.Apply(ctx)
		},
	})
}
