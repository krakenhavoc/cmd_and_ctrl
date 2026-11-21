package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifegift — Enchantment {2}{G} (Betrayers of Kamigawa, #5602):
//
//	"Whenever a land enters, you may gain 1 life."
//
// Any land, any player's: the condition is the land-entered read Zo-Zu
// the Punisher uses (b39LandEnteredController), one trigger per land. The
// "you may" is the trigger's own yes/no as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c99d4be7-244c-4fb5-877a-8aee961a2666",
		Name:         "Lifegift",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B39LandEnteredController(ev, g)
				return ok
			}, "Lifegift — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Lifegift — gain 1 life?"),
		},
	})
}
