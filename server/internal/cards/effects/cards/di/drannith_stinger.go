package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drannith Stinger — Creature — Human Wizard {1}{R}, 2/2 (Jumpstart 2022,
// #5948):
//
//	"Whenever you cycle another card, this creature deals 1 damage to each
//	 opponent.
//	 Cycling {1} ({1}, Discard this card: Draw a card.)"
//
// Cycling is the keyword helper (CR 702.29). The trigger watches EventCycle by
// its controller on any card but the Stinger itself, which is cycled from
// hand and so is never on the battlefield to see it (Basri's watch, from the
// battlefield). One trigger per card cycled, Tal Shiar Analyst's one damage to
// each opponent as one instance.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "edd1727e-3a6e-4d0e-b7ae-fae2108d2147",
		Name:         "Drannith Stinger",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCycle, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventCycle && ev.Actor == source.Controller && ev.CardID != source.InstanceID
			}, "Drannith Stinger — 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1})),
		},
	})
}
