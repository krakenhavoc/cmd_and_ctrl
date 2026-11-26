package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whirlpool Warrior — Creature — Merfolk Warrior {2}{U}, 2/2 (Planechase 2012,
// S82 OLD, #5569):
//
//	"When this creature enters, shuffle the cards from your hand into your
//	 library, then draw that many cards.
//	 {R}, Sacrifice this creature: Each player shuffles the cards from their
//	 hand into their library, then draws that many cards."
//
// Winds of Change's shuffle-and-redraw (each hand is counted as it is put into
// the library, each library shuffled, then every player draws that many): over
// the controller alone on the enters trigger, over every player on the
// sacrifice ability, whose cost is paid before it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7a1c54b-bb2f-4423-865a-ad114cd18c76",
		Name:         "Whirlpool Warrior",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Whirlpool Warrior — shuffle the cards from your hand into your library, then draw that many cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.WindsOfChangeShuffleFrom(g, []uuid.UUID{item.Controller}, nil)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, Sacrifice this creature: Each player shuffles the cards from their hand into their library, then draws that many cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.WindsOfChangeShuffleFrom(g, effects.TablePlayers(effects.NewContext(g, item)), nil)
			},
		}},
	})
}
