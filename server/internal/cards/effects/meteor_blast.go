package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meteor Blast — Sorcery {X}{R}{R}{R} (Commander 2015, #5546):
//
//	"Meteor Blast deals 4 damage to each of X targets."
//
// Crackle with Power's target clause without "up to": TargetAny with
// CountFromX, so the caster names exactly X targets (any target each) and an
// X of 0 names none. Each target still legal as it resolves takes 4 (CR
// 608.2b per slot).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0ba1837b-67a1-40d1-b617-8c94834d02ba",
		Name:         "Meteor Blast",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      c15MeteorBlastTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b10DamageEachLegalTarget(ctx, 4)
		},
	})
}

// c15MeteorBlastTargets is "X targets", any target each.
func c15MeteorBlastTargets() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "X targets"
	spec.CountFromX = true
	return spec
}
