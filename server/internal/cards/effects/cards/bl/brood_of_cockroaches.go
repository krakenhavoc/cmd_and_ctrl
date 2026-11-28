package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brood of Cockroaches — Creature — Insect {1}{B}, 1/1 (Visions, S82 older sets
// batch 791, #5631):
//
//	"When this creature is put into your graveyard from the battlefield, at the
//	 beginning of the next end step, you lose 1 life and return this card to
//	 your hand."
//
// Zodiac Dragon's "put into your graveyard" condition (the card goes to its
// owner's graveyard while that owner also controls it), then a delayed trigger
// (CR 603.7) for the next end step that names the card as it now is: the life
// is lost either way, and the card returns only if it is still that card in
// the graveyard (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a22d694-29d1-4e7b-8fca-d7008aede489",
		Name:         "Brood of Cockroaches",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source) && source.Owner == source.Controller
			}, "Brood of Cockroaches — at the beginning of the next end step, you lose 1 life and return this card to your hand",
				effects.Old791BroodOfCockroaches),
		},
	})
}
