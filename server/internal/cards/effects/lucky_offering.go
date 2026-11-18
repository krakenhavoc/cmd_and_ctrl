package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lucky Offering — Sorcery {W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Destroy target artifact with mana value 3 or less. You gain 3 life."
//
// The artifact's mana value is judged as the spell is cast and again as it
// resolves (CR 608.2b). The life is gained whether or not the destroy
// happened (an indestructible artifact still pays the 3 life) as long as
// the spell resolves at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cbb39e5c-e31a-44e9-be0c-2003c8bce599",
		Name:         "Lucky Offering",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact with mana value 3 or less", And(Artifact(), ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
