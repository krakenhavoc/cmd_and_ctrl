package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charforger — Creature — Phyrexian Beast {1}{B}{R}, 2/3 (Phyrexia: All
// Will Be One):
//
//	"When this creature enters, create a 1/1 red Phyrexian Goblin
//	 creature token.
//	 Whenever another creature or artifact you control is put into a
//	 graveyard from the battlefield, put an oil counter on this creature.
//	 Remove three oil counters from this creature: Exile the top card of
//	 your library. You may play that card this turn."
//
// The counter trigger is Vat of Rebirth's reading. The activation pays
// exactly three oil counters at announcement (it can't be activated with
// fewer), at any time you have priority, and exiles the top card with a
// this-turn permission to play it (a land counts against the land drop).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6ded5b0-1309-4f1e-be10-d84ca0a5dd20",
		Name:         "Charforger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Charforger — create a 1/1 red Phyrexian Goblin creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Phyrexian Goblin"), N: 1})),
			effects.On(game.EventLTB, effects.S82OilOnDeparture, "Charforger — put an oil counter on it",
				effects.Do(effects.CounterOnThis{Kind: effects.S82OneOil, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove three oil counters from this creature: Exile the top card of your library. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.RemoveCountersFromThis(effects.S82OneOil, 3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopWithPermission{
					From: item.Controller, GrantTo: item.Controller, N: 1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
