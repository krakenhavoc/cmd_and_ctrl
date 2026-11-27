package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Recluse — Creature — Ogre Warrior {3}{R}, 5/4 (Betrayers of Kamigawa, S82 OLD
// b595):
//
//	"Whenever a player casts a spell, tap this creature."
//
// One trigger per spell, whoever casts it. It taps on casting, so the creature is tapped
// even if the spell is countered; a tapped Recluse stays tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17ff0155-0006-464a-a65f-b8e1b9aeba0b",
		Name:         "Ogre Recluse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCast
			}, "Ogre Recluse — tap this creature", effects.OldB595TapThis),
		},
	})
}
