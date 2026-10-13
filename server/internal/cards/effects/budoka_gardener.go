package effects

import (
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
const budokaGardenerOracleID = "cb952c81-c7bb-4670-938d-296d35551986"

func init() {
	Register(Spec{
		OracleID:     budokaGardenerOracleID,
		Name:         "Budoka Gardener",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield. If you control ten or more lands, flip this creature.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				put := MayPutALandFromHand("Budoka Gardener")
				put.Then = func(g *game.Game, _ PutFromHandResult) error {
					return flipThisWhen(g, item, controlsAtLeast(10, game.Card.IsLand))
				}
				return put.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:     budokaGardenerOracleID + "#1",
		Name:         "Dokai, Weaver of Life",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{G}, {T}: Create an X/X green Elemental creature token, where X is the number of lands you control.",
			Cost:    Plus(ManaCost("{4}{G}{G}"), TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := countControlled(g, item.Controller, game.Card.IsLand)
				return CreateToken{
					Controller: item.Controller,
					Template: game.Card{
						Name:      "Elemental",
						TypeLine:  "Token Creature — Elemental",
						Colors:    []string{"G"},
						Power:     x,
						Toughness: x,
					},
					N: 1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
