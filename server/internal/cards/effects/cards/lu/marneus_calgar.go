package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marneus Calgar — Legendary Creature — Astartes Warrior {2}{W}{U}{B}, 3/5:
//
//	"Double strike
//	 Master Tactician — Whenever one or more tokens you control enter, draw a
//	 card.
//	 Chapter Master — {6}: Create two 2/2 white Astartes Warrior creature
//	 tokens with vigilance."
//
// The draw is batched: tokens that enter together (the two Warriors) draw one
// card, as "one or more" prints.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6479f7e-01f4-49f1-a444-04bf38934f6b",
		Name:            "Marneus Calgar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsToken()
			}, "Marneus Calgar — draw a card (tokens entered)", effects.Do(effects.DrawCards{N: 1}))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Create two 2/2 white Astartes Warrior creature tokens with vigilance.",
			Cost:    effects.ManaCost("{6}"),
			Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2, Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("2/2 white Astartes Warrior with vigilance"), N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
