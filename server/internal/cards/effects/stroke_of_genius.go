package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stroke of Genius — Instant for {X}{2}{U}:
//
//	"Target player draws X cards."
//
// S20 sub-PR 3: a targeted X spell. Drawing more cards than the
// library holds flags the drawer for the CR 704.5b loss at the next
// SBA check — the classic Stroke kill.
func init() {
	Register(Spec{
		OracleID: "0cc6d683-366f-4ae4-be60-20ad9621fdaf",
		Name:     "Stroke of Genius",
		Targets:  TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: ctx.X()}.Apply(ctx)
		},
	})
}
