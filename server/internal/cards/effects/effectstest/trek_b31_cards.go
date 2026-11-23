package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared helpers ------------------------------------------------------------

// trekB31Enter puts a card from owner's hand onto the battlefield (an
// ordinary entry, so enters triggers fire) and returns its ID.
func TrekB31Enter(t *testing.T, g *game.Game, owner *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = owner.ID, owner.ID
	g.WithWriteLock(func() { owner.Hand.PushTop(c) })
	var id uuid.UUID
	g.WithWriteLock(func() {
		var err error
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{Controller: owner.ID})
		if err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	return id
}

// trekB31Library replaces p's library with the given cards, the first
// one on top.
func TrekB31Library(g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	g.WithWriteLock(func() {
		p.Library.Cards = nil
		for i := len(cards) - 1; i >= 0; i-- {
			c := cards[i]
			if c.InstanceID == uuid.Nil {
				c.InstanceID = uuid.New()
			}
			c.Owner, c.Controller = p.ID, p.ID
			p.Library.PushTop(c)
			ids[i] = c.InstanceID
		}
	})
	return ids
}
