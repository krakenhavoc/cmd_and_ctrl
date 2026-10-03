package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Test of Faith — Instant {1}{W}:
//
//	"Prevent the next 3 damage that would be dealt to target creature this turn. For each 1 damage prevented this way, put a +1/+1 counter on that creature."
//
// ADR 0108 owner decision 2 (#1906): the charged shield (CR 615.7) with a
// CR 615.5 additional effect, owed once per damage instance with what the
// charge prevented. The counters go on as the damage is prevented, before
// the game checks for lethal damage: a 1/1 dealt 6 ends up a 4/4 with 3
// damage marked (the ruling). Damage that can't be prevented puts on none
// and leaves the shield whole (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "3397aa3d-bf73-4ca3-a806-059361603079",
		Name:         "Test of Faith",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return countersPerPreventedShield(item, ctx, 3, "Test of Faith")
		},
	})
}
