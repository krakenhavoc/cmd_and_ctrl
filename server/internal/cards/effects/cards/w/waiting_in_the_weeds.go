package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waiting in the Weeds — Sorcery {1}{G}{G} (Mirage, #5633):
//
//	"Each player creates a 1/1 green Cat creature token for each untapped
//	 Forest they control."
//
// As it resolves each player, starting with the caster, counts the untapped
// Forests they control (anything with the Forest subtype) and creates that
// many Cats.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a61d51e-1a3f-49f6-9a94-a708dd6f69df",
		Name:         "Waiting in the Weeds",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			forest := ctx.TextSubtype("Forest")
			for _, id := range effects.SeatsFromController(ctx.Game, ctx.Controller()) {
				n := effects.CountControlled(ctx.Game, id, func(c game.Card) bool { return c.HasSubtype(forest) && !c.Tapped })
				if n <= 0 {
					continue
				}
				if err := (effects.CreateToken{Controller: id, Template: effects.TokenCard("1/1 green Cat"), N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
