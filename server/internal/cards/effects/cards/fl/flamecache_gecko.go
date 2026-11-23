package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamecache Gecko — Creature — Lizard Warlock {1}{R}, 2/2 (Bloomburrow,
// #4467):
//
//	"When this creature enters, if an opponent lost life this turn, add
//	 {B}{R}.
//	 {1}{R}, Discard a card: Draw a card."
//
// The entry trigger is Bloodtithe Collector's intervening "if" (CR
// 603.4), checked as it would trigger and again as it resolves, on Priest
// of Gix's mana: a triggered ability that adds mana but is not a mana
// ability (CR 605.1b), so it uses the stack and the mana lands as it
// resolves. The rummage is Charging Strifeknight's discard cost (CR
// 601.2h via 602.2b) with a mana cost in place of the tap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d767463b-0fc9-43fc-a85f-e60f030f9b96",
		Name:         "Flamecache Gecko",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Flamecache Gecko — add {B}{R} (an opponent lost life this turn)",
				effects.Do(effects.If{
					Cond: effects.AvB34CondAnOpponentLostLife{},
					Then: []effects.Applier{effects.AddMana{Produced: "{B}{R}"}},
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Discard a card: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
