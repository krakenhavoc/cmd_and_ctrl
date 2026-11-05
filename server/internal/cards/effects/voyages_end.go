package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voyage's End — Instant {1}{U} (Jumpstart 2020):
//
//	"Return target creature to its owner's hand. Scry 1."
//
// Unsummon with a scry after it. The bounce happens first, then the
// scry prompt; the scry is the last instruction, so nothing waits on
// it. A target that is no longer legal as the spell resolves makes the
// whole spell do nothing, the scry included (CR 608.2b).
//
// No simplification. Deck request #4894 (Jumpstart Azorius).
func init() {
	Register(Spec{
		OracleID:     "1c2829b7-f73f-4023-95f3-9cd40b4125d7",
		Name:         "Voyage's End",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
