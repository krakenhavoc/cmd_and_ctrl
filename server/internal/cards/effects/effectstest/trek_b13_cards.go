package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trekB13Card puts a card into a zone of p's: a catalog card when oracle is
// set, with its printed cost and P/T.
func TrekB13Card(g *game.Game, z *game.Zone, p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		z.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// trekB13Cast casts a card strictly with the given params.
func TrekB13Cast(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana string, params game.CastSpellParams) error {
	t.Helper()
	if mana != "" {
		FloatMana(t, g, p, mana)
	}
	params.Strict = true
	return g.CastSpell(p.ID, id, params)
}

// trekB13Tokens is the tokens named `name` that `controller` controls.
func TrekB13Tokens(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if v := LayeredCard(t, g, c.InstanceID); v.IsToken() {
				out = append(out, v)
			}
		}
	}
	return out
}
