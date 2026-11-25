package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Conqueror's Galleon // Conqueror's Foothold — Artifact — Vehicle {4},
// 2/10 (Ixalan, S82 older sets):
//
//	"When this Vehicle attacks, exile it at end of combat, then return it to
//	 the battlefield transformed under your control.
//	 Crew 4"
//
//	Conqueror's Foothold — Land:
//	"{T}: Add {C}.
//	 {2}, {T}: Draw a card, then discard a card.
//	 {4}, {T}: Draw a card.
//	 {6}, {T}: Return target card from your graveyard to your hand."
//
// The attack trigger schedules a delayed trigger for the beginning of the end
// of combat step (CR 511.2) that exiles the Galleon and returns it transformed
// under its controller's control (Ajani Nacatl Pariah's ExileAndReturnTransformed:
// a new object, CR 400.7). If the Galleon has left the battlefield by then,
// or is a new object, nothing happens. The back face is a land with a mana
// ability and three activated abilities.
//
// No simplification.
func init() {
	const oracle = "88b18901-50cd-461c-b1bc-be900210be8e"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Conqueror's Galleon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Conqueror's Galleon — exile it at end of combat, then return it transformed",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ScheduleDelayedTrigger{
						At:    game.StepEndCombat,
						Label: "Conqueror's Galleon — exile it, then return it to the battlefield transformed",
						Body:  conquerorsGalleonBody,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 4",
			Cost:    effects.CrewCost(4),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Conqueror's Galleon"),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Conqueror's Foothold",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.LootOne(g, item, 1)
				},
			},
			{
				Label:   "{4}, {T}: Draw a card.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{6}, {T}: Return target card from your graveyard to your hand.",
				Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost()),
				Targets: effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
				Effect:  effects.ReturnFirstLegalGraveyardTargetToHand,
			},
		},
	})
}

// conquerorsGalleonBody is the delayed trigger's body: the Galleon, if it is
// still the object that attacked, is exiled and returns transformed under its
// controller's control. An on-disk key.
var conquerorsGalleonBody = game.SimpleDelayedBody("conquerors-galleon/exile-and-return-transformed",
	func(g *game.Game, item *game.StackItem) error {
		if g.AbilitySourceGoneForEffect(item) {
			return nil
		}
		return effects.ExileAndReturnTransformed{Target: item.SourceCardID, Controller: item.Controller}.Apply(effects.NewContext(g, item))
	})
