package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stampede Surfer — Creature — Human Warrior {3}{R/G}{R/G}, 4/4 (Ravnica:
// Clue Edition, #5929):
//
//	"Haste
//	 Whenever this creature attacks, for each opponent, you create a 2/2
//	 green Boar creature token that's tapped and attacking that opponent."
//
// One Boar per opponent still in the game, each attacking its own
// opponent (Stangg's AttackOpponent seam); the tokens enter together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22640b75-8db6-4264-a4aa-9174c159a228",
		Name:            "Stampede Surfer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self,
				"Stampede Surfer — for each opponent, create a 2/2 green Boar tapped and attacking that opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					tmpl := ctx.TextToken(effects.TokenCard("2/2 green Boar"))
					var groups []game.TokenGroup
					for _, opp := range effects.B27OtherOpponents(g, item.Controller, uuid.Nil) {
						groups = append(groups, game.TokenGroup{Template: tmpl, Count: 1,
							Entry: game.TokenEntryOptions{Tapped: true}, AttackOpponent: opp})
					}
					if len(groups) == 0 {
						return nil
					}
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller:         item.Controller,
						Source:             item.SourceCardID,
						Groups:             groups,
						Together:           true,
						AttackChoicesLabel: "Stampede Surfer — Boar",
					}, game.KeyedTail{})
				}),
		},
	})
}
