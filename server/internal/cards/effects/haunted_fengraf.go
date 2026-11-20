package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7c6143f3-ad2c-4d7f-9041-aa59f01d8fb7",
		Name:         "Haunted Fengraf",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this land: Return a creature card at random from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ids := graveyardIDs(g, item.Controller, game.Card.IsCreature)
				pick := g.ChooseAtRandomForEffect(randomDraw(NewContext(g, item)), ids, 1)
				if len(pick) == 0 {
					return nil
				}
				return g.ReturnFromGraveyardForEffect(pick[0], game.ZoneHand)
			},
		}},
	})
}
