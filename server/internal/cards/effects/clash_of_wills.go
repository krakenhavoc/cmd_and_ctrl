package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clash of Wills — Instant {X}{U} (Magic Origins, #5549):
//
//	"Counter target spell unless its controller pays {X}."
//
// Logic Knot's counter without delve: X is read off the spell as it
// resolves, and the target's controller is asked to pay that much (CR
// 107.3). A target that left in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "56aec5bd-a8e8-403f-b339-0fc817426428",
		Name:         "Clash of Wills",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return counterTheTargetSpellUnlessPaid("Clash of Wills", "{"+strconv.Itoa(ctx.X())+"}")(ctx.Game, item)
		},
	})
}
