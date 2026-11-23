package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Buy Your Silence — Sorcery {4}{W} (Streets of New Capenna, #5491):
//
//	"Exile target nonland permanent. Its controller creates a Treasure
//	 token."
//
// The controller is read before the exile moves the card (CR 608.2h), as
// Swords to Plowshares reads its life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fd01a30-0c0e-4ac5-b96d-b2a6dad31f0c",
		Name:         "Buy Your Silence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			victim, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: victim, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
