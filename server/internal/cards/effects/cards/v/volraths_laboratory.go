package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volrath's Laboratory — Artifact {5} (Stronghold, S82 OLD b751, #5627):
//
//	"As this artifact enters, choose a color and a creature type.
//	 {5}, {T}: Create a 2/2 creature token of the chosen color and type."
//
// Two as-enters choices (CR 614.12), both prompts queued for the controller
// (ChooseColorAsEnters, ChooseCreatureTypeAsEnters): the colour lands on
// Card.ChosenColor and the type on Card.NamedTribe. The token is a 2/2 creature
// of exactly those, named for its type (CR 111.4), read off the Laboratory as the
// ability resolves; with either choice still unanswered it creates nothing
// (the weaker direction). They are choices made as the artifact entered, not
// words of its text, so a text-changing effect doesn't reach them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2a6e26c-6bd6-4fca-ba9f-d52d01fc03ef",
		Name:         "Volrath's Laboratory",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			if err := effects.ChooseColorAsEnters(game.ColorForBenefit, "Volrath's Laboratory")(card, ctx); err != nil {
				return err
			}
			return effects.ChooseCreatureTypeAsEnters("Volrath's Laboratory")(card, ctx)
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Create a 2/2 creature token of the chosen color and type.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || src.ChosenColor == "" || src.NamedTribe == "" {
					return nil
				}
				token := effects.RuleDefinedToken(game.Card{
					Name:      src.NamedTribe,
					TypeLine:  "Token Creature — " + src.NamedTribe,
					Power:     2,
					Toughness: 2,
					Colors:    []string{src.ChosenColor},
				})
				return g.CreateTokenForEffect(item.Controller, token, 1)
			},
		}},
	})
}
