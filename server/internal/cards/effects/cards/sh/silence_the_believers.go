package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silence the Believers — Instant {2}{B}{B} (Journey into Nyx):
//
//	"Strive — This spell costs {2}{B} more to cast for each target beyond the first.
//	 Exile any number of target creatures and all Auras attached to them."
//
// The creatures and the Auras attached to them (read as the spell resolves)
// leave together as one exile event.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c2de008-5747-49c0-89c1-83331cb55896",
		Name:         "Silence the Believers",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{2}{B}", "Strive — This spell costs {2}{B} more to cast for each target beyond the first."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ids := effects.S82OldB366WithAttachedAuras(ctx, effects.S82OldB366LegalTargetCards(ctx))
			if len(ids) > 0 {
				ctx.Game.ExileCardsForEffect(ids)
			}
			return nil
		},
	})
}
