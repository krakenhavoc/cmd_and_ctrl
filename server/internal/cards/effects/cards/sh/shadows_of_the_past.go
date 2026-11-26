package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadows of the Past — Enchantment {1}{B} (Magic Origins, S82 older sets):
//
//	"Whenever a creature dies, scry 1.
//	 {4}{B}: Each opponent loses 2 life and you gain 2 life. Activate only if
//	 there are four or more creature cards in your graveyard."
//
// The trigger is Ashes of the Abhorrent's, one for every creature that dies,
// yours or not, with a scry 1 for its controller. The ability is Gate to the
// Afterlife's CR 602.5b "activate only if" over the creature cards in your
// graveyard (GraveyardAtLeast), with Sol'Kanar the Tainted's drain: life loss
// from each opponent, and one gain of 2 however many opponents there are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5ba544f-b223-47c8-a408-b00512772107",
		Name:         "Shadows of the Past",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			}, "Shadows of the Past — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{4}{B}: Each opponent loses 2 life and you gain 2 life. Activate only if there are four or more creature cards in your graveyard.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{4}{B}"),
			Condition: effects.GraveyardAtLeast(4, func(c game.Card) bool { return c.IsCreature() }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachOpponentLosesTwoYouGainTwo(item, effects.NewContext(g, item), 0)
			},
		}},
	})
}
