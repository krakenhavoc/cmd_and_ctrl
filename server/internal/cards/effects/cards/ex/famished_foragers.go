package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Famished Foragers — Creature — Vampire {3}{R}, 4/3 (Innistrad: Midnight
// Hunt, #5497):
//
//	"When this creature enters, if an opponent lost life this turn, add
//	 {R}{R}{R}.
//	 {2}{R}, Discard a card: Draw a card."
//
// Flamecache Gecko's entry mana and rummage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e95f22dd-b89a-4836-86ad-5a9a1627af38",
		Name:         "Famished Foragers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Famished Foragers — add {R}{R}{R} (an opponent lost life this turn)",
				effects.Do(effects.If{
					Cond: effects.AvB34CondAnOpponentLostLife{},
					Then: []effects.Applier{effects.AddMana{Produced: "{R}{R}{R}"}},
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Discard a card: Draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.DiscardACard()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
