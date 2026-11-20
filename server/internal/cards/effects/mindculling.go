package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindculling — Sorcery {5}{U} (New Phyrexia, #5575):
//
//	"You draw two cards and target opponent discards two cards."
//
// The two draws happen first, then the target opponent is asked to
// discard two (targetOpponentDiscards), in printed order (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "955a7f4d-7100-4d8b-818c-be8651d0ebe6",
		Name:         "Mindculling",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return targetOpponentDiscards(2)(ctx.Game, item)
		},
	})
}
