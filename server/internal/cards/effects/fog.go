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
// Turn-scoped rather than battlefield-presence-gated because Fog is
// an instant: it resolves, registers a transient replacement, and
// goes to the graveyard. The pipeline keeps firing the registered
// replacement on every combat-damage event until StepCleanup clears
// the turn-scoped registry — which is exactly CR 615.6's "this
// turn".
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
