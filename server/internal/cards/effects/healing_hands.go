package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Healing Hands — Sorcery {2}{W} (Magic Origins, #5549):
//
//	"Target player gains 4 life.
//	 Draw a card."
//
// Sylvan Bounty's life gain, then a draw for the caster.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cc48835-3ac0-4774-b380-f9b21d2dc974",
		Name:         "Healing Hands",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 4}),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := targetPlayerGainsLife(4)(item, ctx); err != nil {
				return err
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}
