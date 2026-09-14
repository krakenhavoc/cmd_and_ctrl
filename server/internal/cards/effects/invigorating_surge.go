package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invigorating Surge — Instant {2}{G} (EDHREC rank 3561):
//
//	"Put a +1/+1 counter on target creature you control, then double
//	 the number of +1/+1 counters on that creature."
//
// The counters pump. Two placements in printed order — one, then as
// many as the creature now carries — so a bare creature ends on two
// and one with three ends on eight, and a Hardened Scales sees both
// placements. A target that left in response fizzles the spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5daac63a-1534-4194-8cb7-506508e364f4",
		Name:         "Invigorating Surge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b34PutCounterThenDoubleCounters(ctx)
		},
	})
}
