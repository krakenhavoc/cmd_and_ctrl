package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shamble Back — Sorcery {B} (Shadows over Innistrad, #5543):
//
//	"Exile target creature card from a graveyard. Create a 2/2 black Zombie
//	 creature token. You gain 2 life."
//
// Vile Rebirth's exile and Zombie with the life gain: a creature card in any
// graveyard. With its one target gone the spell does nothing, token and life
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11a20648-fdb6-4f52-8fda-c66985e0c044",
		Name:         "Shamble Back",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerGraveyard},
		Targets:      effects.TargetCreatureInAnyGraveyard(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.CreateToken{Controller: item.Controller, Template: effects.BlackZombieToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
