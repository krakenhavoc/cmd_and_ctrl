package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wallop — Sorcery {1}{G} (Invasion, #5616):
//
//	"Destroy target blue or black creature with flying."
//
// One target clause judged at announce and again at resolution (CR 608.2b):
// colour and flying are read through the layered view, so a creature that
// lost flying in response is no longer a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e939ef5e-ebe1-48cd-9e9f-eb3fd516165c",
		Name:         "Wallop",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target blue or black creature with flying", effects.Or(effects.OfColor("U"), effects.OfColor("B")), effects.HasKeyword("flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
