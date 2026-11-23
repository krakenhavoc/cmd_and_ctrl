package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warbeast of Gorgoroth — Creature — Beast {4}{R}, 5/4 (The Lord of the
// Rings: Tales of Middle-earth Commander):
//
//	"Whenever this creature or another creature you control with power 4 or
//	 greater dies, amass Orcs 2."
//
// A died-under-your-control watch (Buzzard-Wasp Colony's "another creature
// you control dies", plus this creature itself) whose power test reads the
// creature as it last stood on the battlefield (CR 603.10a, 608.2h):
// the power it had with every pump and counter, not the one printed. Foray
// of Orcs's amass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ca52e15-5a55-4d79-b0bf-0e2dc515ae6a",
		Name:         "Warbeast of Gorgoroth",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				if !ok || effects.LeftUnderControlOf(ev, dead) != source.Controller {
					return false
				}
				return effects.B29PowerAsItLastStood(g, dead.InstanceID) >= 4
			}, "Warbeast of Gorgoroth — amass Orcs 2", effects.Do(effects.Amass{Subtype: "Orc", N: 2})),
		},
	})
}
