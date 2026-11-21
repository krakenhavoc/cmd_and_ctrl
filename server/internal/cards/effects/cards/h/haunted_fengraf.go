package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haunted Fengraf — Land (Dark Ascension, #5571):
//
//	"{T}: Add {C}.
//	 {3}, {T}, Sacrifice this land: Return a creature card at random
//	 from your graveyard to your hand."
//
// The card is drawn from the keyed random stream as the ability resolves
// (Surreal Memoir's pick), so an undo replays it; with no creature card in
// the graveyard nothing returns.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c6143f3-ad2c-4d7f-9041-aa59f01d8fb7",
		Name:         "Haunted Fengraf",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this land: Return a creature card at random from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ids := effects.GraveyardIDs(g, item.Controller, game.Card.IsCreature)
				pick := g.ChooseAtRandomForEffect(effects.RandomDraw(effects.NewContext(g, item)), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
			},
		}},
	})
}
