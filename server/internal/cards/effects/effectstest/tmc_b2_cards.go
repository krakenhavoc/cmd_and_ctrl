package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmcB2Cast puts `c` in p's hand and casts it with a strict payment
// from p's pool, returning the card's id and the cast's error.
func TmcB2Cast(g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	params.Strict = true
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// tmcB2Pool adds mana to p's pool.
func TmcB2Pool(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("add %s: %v", mana, err)
	}
}
