package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyroceratops — Creature — Elemental Dinosaur {3}{R}, 2/3 (Ikoria: Lair of Behemoths, #5512):
//
//	"Trample
//	 Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	 creature."
//
// The spell's type is read where it sits on the stack, so a creature card
// cast as something else is judged by what it is as cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5306bf4-0cba-48dd-8af8-7c7821db7e02",
		Name:            "Pyroceratops",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.B10NoncreatureSpellCastByYou, "Pyroceratops — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
