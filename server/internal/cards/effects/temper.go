package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temper — Instant {X}{1}{W}:
//
//	"Prevent the next X damage that would be dealt to target creature this turn. For each 1 damage prevented this way, put a +1/+1 counter on that creature."
//
// ADR 0108 owner decision 2 (#1906): Test of Faith with X for the charge
// (CR 615.7). The CR 615.5 additional effect is owed once per damage
// instance with what the charge prevented, so X larger than the damage
// gives only one counter per point actually prevented (the ruling), and
// the counters go on with any unprevented damage, before the game checks
// for lethal damage. Damage that can't be prevented puts on none and
// leaves the shield whole (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9cefff18-96b1-4090-93cc-1ca981a58fbd",
		Name:         "Temper",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return countersPerPreventedShield(item, ctx, ctx.X(), "Temper")
		},
	})
}
