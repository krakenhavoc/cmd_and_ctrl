package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crackle with Power — Sorcery {X}{X}{X}{R}{R} (EDHREC rank 1184):
//
//	"Crackle with Power deals five times X damage to each of up to X
//	 targets."
//
// The red deck's finisher: three X pips, and the count of targets is
// X itself. The cost engine charges 3X generic (ParseCost counts
// every {X}); the target clause is TargetAny with CountFromX, the
// Waterbender's Restoration hook that replaces the clause's count
// with the announced X before anything validates against it — so an
// X of 0 buys no targets rather than an unbounded clause. Each still-
// legal target takes 5X from the spell (CR 608.2b per slot).
//
// "Up to X targets" is UpToX: the announced X is the ceiling, so a
// caster may choose fewer targets than X and still deal five times X
// to each (#1738).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "273f5483-b67e-4dd6-bba8-c0a047fa34d7",
		Name:         "Crackle with Power",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      b10CrackleTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b10DamageEachLegalTarget(ctx, 5*ctx.X())
		},
	})
}

// b10CrackleTargets is "up to X targets", any target each.
func b10CrackleTargets() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "up to X targets"
	spec.CountFromX = true
	spec.UpToX = true
	return spec
}
