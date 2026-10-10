package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestXCeilingViewCarriesTheCastersCeiling is the #2581 view half: a
// spell with a printed "X can't be greater than <count>" carries the
// count CastSpell refuses above as `x_max`, read for the card's caster,
// present at 0; a card with no ceiling carries none.
func TestXCeilingViewCarriesTheCastersCeiling(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-x-ceiling"
	prev := game.CatalogXCeiling
	game.CatalogXCeiling = func(key string) *game.XCeiling {
		if key == oracle {
			return &game.XCeiling{
				Label: "the number of creatures you control",
				Count: func(g *game.Game, caster uuid.UUID) int {
					return g.CountControlledMatchingForEffect(caster, game.PermanentQuery{Types: []string{"creature"}})
				},
			}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogXCeiling = prev })
	id, plain := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Ceiling Thing", TypeLine: "Sorcery", OracleID: oracle,
			ManaCost: "{X}{U}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		me.Hand.PushTop(game.Card{
			InstanceID: plain, Name: "Plain X", TypeLine: "Sorcery", OracleID: "test-view-no-ceiling",
			ManaCost: "{X}{U}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	if c.XMax == nil || *c.XMax != 0 {
		t.Fatalf("no creature: x_max = %v, want present and 0", c.XMax)
	}
	if p := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), plain.String()); p.XMax != nil {
		t.Errorf("a card with no ceiling carries x_max %d", *p.XMax)
	}

	pushViewCreature(g, me.ID, 1, false)
	pushViewCreature(g, me.ID, 1, false)
	pushViewCreature(g, g.Seats[1].ID, 1, false)
	c = handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), id.String())
	var want int
	g.WithWriteLock(func() { want, _ = g.SpellXCeilingForEffect(me.ID, oracle) })
	if want != 2 || c.XMax == nil || *c.XMax != want {
		t.Errorf("two creatures: x_max = %v (engine %d), want 2", c.XMax, want)
	}
}
