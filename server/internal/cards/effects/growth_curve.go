package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Growth Curve — Sorcery {G}{U} (Secrets of Strixhaven, #4456):
//
//	"Put a +1/+1 counter on target creature you control, then double the
//	 number of +1/+1 counters on that creature."
//
// Invigorating Surge's body (b34PutCounterThenDoubleCounters) at sorcery
// speed: two placements in printed order, one, then as many as the creature
// now carries, so a bare creature ends on two and one with three ends on
// eight, and a Hardened Scales sees both placements. A target that left in
// response fizzles the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1efd6082-15a8-4322-892d-02d01e381e2a",
		Name:         "Growth Curve",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34PutCounterThenDoubleCounters(ctx)
		},
	})
}
