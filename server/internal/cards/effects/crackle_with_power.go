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
// Sandbox simplification, declared: "up to X targets" is EXACTLY X
// targets. CountFromX pins both bounds to X, and there is no
// X-bounded "up to" shape. A caster who wants fewer targets than X
// announces a smaller X — weaker than printed (less damage), never
// stronger.
func init() {
	Register(Spec{
		OracleID:     "273f5483-b67e-4dd6-bba8-c0a047fa34d7",
		Name:         "Crackle with Power",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You must choose exactly X targets rather than up to X."},
		Targets:      b10CrackleTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b10DamageEachLegalTarget(ctx, 5*ctx.X())
		},
	})
}

// b10CrackleTargets is "X targets", any target each.
func b10CrackleTargets() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "X targets"
	spec.CountFromX = true
	return spec
}
