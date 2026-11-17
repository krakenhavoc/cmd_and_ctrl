package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infiltrate — Instant {U}
// (Nemesis, #5618):
//
//	"Target creature can't be blocked this turn."
//
// Artful Dodge without flashback: the shared
// targetCantBeBlockedThisTurn body, ending in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "081764b4-ef96-44aa-836c-05353efd215c",
		Name:         "Infiltrate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return targetCantBeBlockedThisTurn("Infiltrate")(ctx.Game, item)
		},
	})
}
