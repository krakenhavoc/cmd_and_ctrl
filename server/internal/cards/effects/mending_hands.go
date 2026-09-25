package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mending Hands — Instant {W}:
//
//	"Prevent the next 4 damage that would be dealt to any target
//	 this turn."
//
// The charged shield, and the reason PreventNextDamage exists. One
// clause, no rider, no life gain — every other card in the family
// (Candles' Glow, Test of Faith, Divine Deflection) is this plus a
// second sentence, so this is the one that pins the arithmetic
// without anything else in the way.
//
// The arithmetic is the point. Four points of shield against six
// damage prevents four and lets two through, and the shield is then
// empty; against three it prevents all three and keeps one for the
// next event. "Cancel the event whenever the shield covers any of
// it" is the natural first draft and is strictly stronger than
// printed.
//
// "Any target" is the modern wording for "creature, player or
// planeswalker", which is TargetAny() — and the shield does not
// care which it got, because ReplacementEvent.DamageTarget is one
// UUID either way.
//
// Note the shield outlives the spell: Mending Hands is in the
// graveyard before the damage it prevents is dealt. That is the
// same until-end-of-turn ScopedEffect lifetime Fog uses, and it is
// why neither card needs a permanent to hang off. The charge left is
// part of the record, so an undo rewinds a spent shield.
func init() {
	Register(Spec{
		OracleID:     "a612f30d-cd55-438b-a7de-8c80509183aa",
		Name:         "Mending Hands",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return PreventNextDamage{
				Target: item.Targets[0].ID,
				Amount: 4,
				Label:  "Mending Hands: prevent the next 4 damage",
			}.Apply(ctx)
		},
	})
}
