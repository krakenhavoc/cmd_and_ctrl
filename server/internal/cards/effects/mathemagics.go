package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mathemagics — Sorcery {X}{X}{U}{U} (Secrets of Strixhaven, #4456):
//
//	"Target player draws 2ˣ cards. (2⁰ = 1, 2¹ = 2, 2² = 4, 2³ = 8,
//	 2⁴ = 16, 2⁵ = 32, and so on.)"
//
// X is announced once and paid twice: the cost engine counts every {X}
// pip, as it does for Crackle with Power. The count is 2 to the X, so
// X = 0 draws one card. It is held to the table's count ceiling
// (game.MaxCount) so a huge X cannot overflow; the draws stop when the
// library runs out, and the player who drew from an empty library loses
// at the next state-based check (CR 704.5b), as Stroke of Genius does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f56a303-b3da-41df-a959-7823a87ee60c",
		Name:         "Mathemagics",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return DrawCards{Player: item.Targets[0].ID, N: mathemagicsCount(ctx.X())}.Apply(ctx)
		},
	})
}

// mathemagicsCount is 2 to the x, pinned to game.MaxCount.
func mathemagicsCount(x int) int {
	if x < 0 {
		return 0
	}
	n := 1
	for i := 0; i < x; i++ {
		n = game.SaturatingMul(n, 2)
		if n == game.MaxCount {
			break
		}
	}
	return n
}
