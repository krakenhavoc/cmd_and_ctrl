package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sbondHand puts a catalog creature card in a player's hand.
func SbondHand(p *game.Player, name, typeLine, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// sbondCast casts a creature from hand and resolves it, leaving whatever
// it triggered on the stack.
func SbondCast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	NghcPassOnce(t, g)
}

// sbondPartner reads who a permanent is paired with.
func SbondPartner(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() { out = g.PairedWithForEffect(id) })
	return out
}
