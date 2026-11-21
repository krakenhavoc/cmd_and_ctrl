package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trading Post — Artifact {4} (EDHREC rank 2140):
//
//	"{1}, {T}, Discard a card: You gain 4 life.
//	 {1}, {T}, Pay 1 life: Create a 0/1 white Goat creature token.
//	 {1}, {T}, Sacrifice a creature: Return target artifact card from
//	 your graveyard to your hand.
//	 {1}, {T}, Sacrifice an artifact: Draw a card."
//
// The artifact deck's Swiss-army knife. All four abilities
// are CR 602 activated abilities built from the cost components the
// engine has — mana, tap, discard-a-card, life, sacrifice-a-creature (the Goat it
// just made qualifies, as printed) and sacrifice-an-artifact (the
// Post itself qualifies, as printed). The regrowth targets an
// artifact card the controller owns, picked in the zone browser and
// re-checked at resolution.
func init() {
	postCost := effects.Plus(effects.ManaCost("{1}"), effects.TapCost())
	effects.Register(effects.Spec{
		OracleID:     "63788566-e25a-44bb-bb55-197e1b93b3e8",
		Name:         "Trading Post",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}, Discard a card: You gain 4 life",
				Purpose: game.Purpose{LifeGain: 4, Answers: game.AnswerValue},
				Cost:    effects.Plus(postCost, effects.DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}, Pay 1 life: Create a 0/1 white Goat",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(postCost, effects.PayLife(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("0/1 white Goat"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}, Sacrifice a creature: Return target artifact card from your graveyard to your hand",
				Cost:    effects.Plus(postCost, effects.SacrificeACreature()),
				Targets: effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.Artifact(), effects.YouOwn()),
				Effect:  effects.S82Mom16ReturnTargetToHand,
			},
			{
				Label:   "{1}, {T}, Sacrifice an artifact: Draw a card",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(postCost, effects.B10SacrificeAnArtifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
