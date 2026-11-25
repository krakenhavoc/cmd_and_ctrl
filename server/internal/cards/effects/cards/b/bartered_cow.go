package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bartered Cow — Creature — Ox {3}{W}, 3/3 (Throne of Eldraine, S82 old
// sets):
//
//	"When this creature dies and when you discard this card, create a Food
//	 token."
//
// One printed ability with two conditions (CR 603.2): it dies (the
// battlefield look-back, CR 603.10a), or its owner discards it from their
// hand. The discarded card has left the hand by the time the trigger is
// looked at, so the ability watches from every zone a discard sends a card
// to, as the "when you discard this card" family does (WhereverDiscarded).
// The engine reads a battlefield exit through the look-back (CR 603.10a)
// and a discard through the declared-zone walk, and a row declaring both
// would fire twice on a death (the card sits in the graveyard for both
// readers), so the one printed sentence is two rows with the same body.
//
// No simplification.
func init() {
	body := effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})
	dies := effects.WhenThisDies("Bartered Cow — create a Food token", body)
	discarded := effects.WhereverDiscarded(effects.On(game.EventDiscardCard,
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return ev.CardID == source.InstanceID && ev.Actor == source.Owner
		},
		"Bartered Cow — you discarded this card: create a Food token", body))
	effects.Register(effects.Spec{
		OracleID:     "33d42811-686d-4aca-bb3b-086de453ddaf",
		Name:         "Bartered Cow",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{dies, discarded},
	})
}
