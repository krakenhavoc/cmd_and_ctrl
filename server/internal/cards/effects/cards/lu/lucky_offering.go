package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cbb39e5c-e31a-44e9-be0c-2003c8bce599",
		Name:         "Lucky Offering",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact with mana value 3 or less", effects.And(effects.Artifact(), effects.ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
