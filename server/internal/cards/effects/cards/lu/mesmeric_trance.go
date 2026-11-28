package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mesmeric Trance — Enchantment {1}{U}{U} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep {1}
//	 {U}, Discard a card: Draw a card."
//
// Anje Falkenrath's rummage: the discard is a cost paid as the ability is activated (CR 118.8),
// so a spell can't be responded with to take the card back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1cc6979-d786-4035-963b-768ac8cef116",
		Name:         "Mesmeric Trance",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.CumulativeUpkeep("Mesmeric Trance — cumulative upkeep {1}", "{1}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Discard a card: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.DiscardACard()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
