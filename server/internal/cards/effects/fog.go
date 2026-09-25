package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fog — "Prevent all combat damage that would be dealt this turn."
//
// The S17 sub-PR 5 proof of concept for damage prevention via the
// CR 614 replacement pipeline, and since S30 the first caller of
// the PreventAllCombatDamageThisTurn primitive its inline closure
// became. The behaviour is unchanged to the byte; what moved is
// where the closure lives, so Holy Day and Tangle are three lines
// each instead of three copies of the same registration.
//
// A duration rather than battlefield presence because Fog is an
// instant: it resolves, registers a replacement effect, and goes to
// the graveyard. The pipeline keeps firing it on every combat-damage
// event until the cleanup step's duration sweep ends it — which is
// exactly CR 514.2's "this turn". The effect is a ScopedEffect record
// (ADR 0041 phase 3 tier 3b), so a table with a Fog on it is still a
// restore point.
//
// Combat damage only. A Lightning Bolt cast after the Fog resolves
// still kills; the IsCombatDamage flag comes from the combat-damage
// resolver and from nothing else.
//
// The charged variant of prevention — "prevent the next N damage",
// which has to track how much it has absorbed — ships alongside in
// S30 as PreventNextDamage. Mending Hands is its first caller.
func init() {
	Register(Spec{
		OracleID:     "27e9db49-7af7-4bef-ad4c-bf5dfb92030d",
		Name:         "Fog",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PreventAllCombatDamageThisTurn{
				Label: "Fog: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
