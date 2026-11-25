package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drannith Healer — Creature — Human Cleric {1}{W}, 2/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever you cycle another card, you gain 1 life.
//	 Cycling {1}"
//
// Cycling is an activated ability from the hand, so the Healer watches only
// other cards being cycled; the trigger still refuses an event naming itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13e2a2c3-5ab5-4de4-96ca-5f9c72137000",
		Name:         "Drannith Healer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.Actor == source.Controller && ev.CardID != source.InstanceID
			}, "Drannith Healer — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
