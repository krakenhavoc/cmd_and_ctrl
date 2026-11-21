package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func NsPromises(p *game.Player) []game.NextSpellPromise {
	var out []game.NextSpellPromise
	for _, s := range p.Statics {
		if s.NextSpell.Active {
			out = append(out, s.NextSpell)
		}
	}
	return out
}

// nsHandCard seeds a card into a seat's hand for a direct CastSpell.
func NsHandCard(p *game.Player, name, typeLine, cost string) game.Card {
	c := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID, Power: 2, Toughness: 2}
	p.Hand.PushTop(c)
	return c
}

// nsPrice prices `card` for `p` through the one CR 601.2f pass every
// cast uses.
func NsPrice(t *testing.T, g *game.Game, p *game.Player, card game.Card) game.ParsedCost {
	t.Helper()
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatal(err)
	}
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: p.ID, FromZone: game.ZoneHand})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
