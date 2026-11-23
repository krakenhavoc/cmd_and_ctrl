package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dstHand puts a card with a mana cost, and the given colours, in a
// player's hand.
func DstHand(p *game.Player, name, typeLine, manaCost string, colors ...string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine, c.ManaCost, c.Colors = typeLine, manaCost, colors
	p.Hand.PushTop(c)
	return c.InstanceID
}

// dstToken creates one token for controller and returns its ID.
func DstToken(t *testing.T, g *game.Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(controller, game.Card{Name: name, TypeLine: typeLine, Power: 1, Toughness: 1}, 1); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	})
	return BattlefieldIDNamed(g, controller, name)
}
