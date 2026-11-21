package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crack Open — Sorcery {2}{G}:
//
//	"Destroy target artifact or enchantment. Create a Treasure token."
//
// The Treasure is made as long as the spell resolves with its target still
// legal (CR 608.2b: with its only target illegal the spell does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "412197f2-0e95-47fb-83db-ae8301c90ff0",
		Name:         "Crack Open",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Tokens: 1},
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
