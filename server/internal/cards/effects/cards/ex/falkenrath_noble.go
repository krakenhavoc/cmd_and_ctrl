package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Noble — Creature — Vampire Noble {3}{B}, 2/2 (EDHREC
// rank 1990):
//
//	"Flying
//	 Whenever this creature or another creature dies, target player
//	 loses 1 life and you gain 1 life."
//
// Blood Artist with wings and two more mana. The same trigger: its
// own death included, any player's creature, one targeted drain per
// death — tedious across a board wipe, and correct.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3739b179-bc81-4737-8376-66a57e16b942",
		Name:            "Falkenrath Noble",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.DiedCreature(ev, g)
				return ok
			},
			Targets: effects.TargetPlayer("target player"),
			Key:     "Falkenrath Noble — drain 1",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
