package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulseExileCost exiles a sorcery with the given mana cost off the
// active seat's library with "you may play it this turn".
func ImpulseExileCost(t *testing.T, g *game.Game, name, cost string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if _, err := g.ExileTopWithPermissionForEffect(me.ID, me.ID, 1, game.CastPermission{}); err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})
	if !g.Exile.Contains(id) {
		t.Fatal("setup: the card is not in exile")
	}
	return me, id
}

func HandSorcery(me *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}
