package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoot Down — Sorcery for {3}{G} (The Brothers' War):
//
//	"Exile target artifact, enchantment, or creature with flying."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a89feeb-9e66-4539-bd75-7d9d53f52d49",
		Name:         "Shoot Down",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target artifact, enchantment, or creature with flying",
			effects.Or(effects.Artifact(), effects.Enchantment(), effects.And(effects.Creature(), effects.HasKeyword("flying")))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			return effects.ExileTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
