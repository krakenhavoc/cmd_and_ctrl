package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6238: Booby Trap's "When the chosen player draws a card with the
// chosen name" reads the drawer and the drawn card's name.
func TestChosenPlayerDrawsChosenNameReadsDrawerAndName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them, other := g.Seats[0], g.Seats[1], g.Seats[2]
	named := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: them.ID, Controller: them.ID}
	plain := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: them.ID, Controller: them.ID}
	theirs := game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: other.ID, Controller: other.ID}
	g.WithWriteLock(func() {
		them.Hand.PushTop(named)
		them.Hand.PushTop(plain)
		other.Hand.PushTop(theirs)
	})
	trap := &game.Card{InstanceID: uuid.New(), Name: "Booby Trap", Controller: me.ID, Owner: me.ID, ChosenPlayer: them.ID, ChosenName: "Lightning Bolt"}
	when := effects.ChosenPlayerDrawsChosenName()
	draw := func(actor, card uuid.UUID) game.Event {
		return game.Event{Kind: game.EventDrawCard, Actor: actor, CardID: card}
	}
	if !when(draw(them.ID, named.InstanceID), trap, game.Characteristic{}, g) {
		t.Error("the chosen player drawing the named card did not trigger")
	}
	if when(draw(them.ID, plain.InstanceID), trap, game.Characteristic{}, g) {
		t.Error("another name triggered")
	}
	if when(draw(other.ID, theirs.InstanceID), trap, game.Characteristic{}, g) {
		t.Error("a player who is not chosen drawing the named card triggered")
	}
	trap.ChosenName = ""
	if when(draw(them.ID, named.InstanceID), trap, game.Characteristic{}, g) {
		t.Error("no name chosen, yet it triggered")
	}
}
