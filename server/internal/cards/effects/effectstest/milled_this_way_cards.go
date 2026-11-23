package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// milled_this_way_cards_test.go — #893 at the catalog level.
//
// The engine half (game/milled_this_way_test.go) taught the mill to
// report what LANDED in the zone it aimed at; this is the same rule on
// the cards that read the mill's answer. It is not the answer they used
// to get: the old list held every leg that had not paused, so a card a
// graveyard replacement sent to exile, and a commander whose owner was
// still being ASKED about the command zone, both counted.
//
// CR 400.7: the object that ARRIVED is the one the effect moved.
//
// ADR 0115: a commander no longer pauses a mill or an exile. It arrives
// like any other card, so it counts, and CR 903.9a asks its owner about
// the command zone afterwards.

// libraryCardFor pushes a card onto the top of p's library. The top is
// the LAST element, so the last push is the first card milled.
func LibraryCardFor(p *game.Player, name, typeLine string, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}
