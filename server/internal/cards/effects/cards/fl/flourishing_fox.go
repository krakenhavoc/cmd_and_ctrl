package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flourishing Fox — Creature — Fox {W}, 1/1 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever you cycle another card, put a +1/+1 counter on this creature.
//	 Cycling {1}"
//
// Cycling is an activated ability from the hand, so the Fox is on the
// battlefield only when it watches another card being cycled; the trigger
// still refuses an event naming the Fox itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44add927-4e6b-408e-907f-6a665add8b16",
		Name:         "Flourishing Fox",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.Actor == source.Controller && ev.CardID != source.InstanceID
			}, "Flourishing Fox — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
