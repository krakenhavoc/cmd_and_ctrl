package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magmaroth — Creature — Elemental {3}{R}, 5/5 (Hour of Devastation, #5534):
//
//	"At the beginning of your upkeep, put a -1/-1 counter on this creature.
//	 Whenever you cast a noncreature spell, remove a -1/-1 counter from this
//	 creature."
//
// Two plain triggers on the Magmaroth's own counters. A noncreature spell is
// any spell without the creature type (a land is not cast). With no -1/-1
// counter on it the removal does nothing (CR 609.3), and a Magmaroth that has
// left cannot be given or lose one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac5c2069-0706-4b85-a811-a522db016e3e",
		Name:         "Magmaroth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Magmaroth — put a -1/-1 counter on it", effects.PutACounterOnThis(game.CounterMinusOne)),
			effects.WheneverYouCast(effects.Noncreature(), "Magmaroth — remove a -1/-1 counter from it", effects.RemoveACounterFromThis(game.CounterMinusOne)),
		},
	})
}
