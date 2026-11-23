package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Budoka Gardener // Dokai, Weaver of Life — {1}{G} Creature — Human
// Monk 2/1 // Legendary Creature — Human Monk 3/3, a flip card (#3216,
// ADR 0151):
//
//	Top:     "{T}: You may put a land card from your hand onto the
//	          battlefield. If you control ten or more lands, flip this
//	          creature."
//	Flipped: "{4}{G}{G}, {T}: Create an X/X green Elemental creature
//	          token, where X is the number of lands you control."
//
// The land drop is the shared MayPutALandFromHand clause. The "if you
// control ten or more lands" is asked after the pick has been answered
// and the land has arrived (the clause's Then), and it is asked even if
// you declined or had no land to put: it is a separate instruction. X
// is counted as the ability resolves. Dokai is green and {1}{G}
// (CR 710.1c).
//
// No simplification.
// budokaGardenerFlip is "if you control ten or more lands, flip this
// creature" as a keyed tail (#4253 slice 6). An on-disk key.
var budokaGardenerFlip = effects.PutFromHandTail("budoka-gardener/flip-with-ten-lands", func(ctx *effects.Context, _ effects.PutFromHandResult) error {
	return effects.FlipThisWhen(ctx.Game, ctx.Item, effects.YouControlAtLeast(10, game.Card.IsLand))
})

const BudokaGardenerOracleID = "cb952c81-c7bb-4670-938d-296d35551986"

func init() {
	effects.Register(effects.Spec{
		OracleID:     BudokaGardenerOracleID,
		Name:         "Budoka Gardener",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield. If you control ten or more lands, flip this creature.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				put := effects.MayPutALandFromHand("Budoka Gardener")
				put.ThenTail = budokaGardenerFlip
				put.ThenItem = true
				return put.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     BudokaGardenerOracleID + "#1",
		Name:         "Dokai, Weaver of Life",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}{G}, {T}: Create an X/X green Elemental creature token, where X is the number of lands you control.",
			Cost:    effects.Plus(effects.ManaCost("{4}{G}{G}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := effects.CountControlled(g, item.Controller, game.Card.IsLand)
				return effects.CreateToken{
					Controller: item.Controller,
					Template: game.Card{
						Name:      "Elemental",
						TypeLine:  "Token Creature — Elemental",
						Colors:    []string{"G"},
						Power:     x,
						Toughness: x,
					},
					N: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
