package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5f56a303-b3da-41df-a959-7823a87ee60c",
		Name:         "Mathemagics",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return effects.DrawCards{Player: item.Targets[0].ID, N: MathemagicsCount(ctx.X())}.Apply(ctx)
		},
	})
}

// mathemagicsCount is 2 to the x, pinned to game.MaxCount.
func MathemagicsCount(x int) int {
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
