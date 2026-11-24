package testsl

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

const laviniaAzoriusRenegadeOracle = "81ba3a33-0e04-4f68-9a42-1dd717733b57"

// laviniaBoard puts Lavinia under seat 0 and gives seat 1, the active
// player in its main phase, two Forests and six green mana.
func laviniaBoard(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Lavinia, Azorius Renegade", "Legendary Creature — Human Soldier", laviniaAzoriusRenegadeOracle, 2, 2)
	effectstest.AangAdvanceToMain(t, g, 1)
	g.WithWriteLock(func() {
		effectstest.PushLands(g, opp.ID, 2)
		for range 6 {
			opp.ManaPool.AddMana(game.ManaToken{Color: "G"})
		}
	})
	return g, me, opp
}

func laviniaHandCard(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// An opponent with two lands can't cast a noncreature spell of mana
// value 3, nor an X spell whose X makes it 3 (CR 202.3e); a creature
// spell of any size, a spell within the count, and Lavinia's controller
// are not restricted.
func TestLaviniaAzoriusRenegadeCapsOpponentsNoncreatureSpells(t *testing.T) {
	g, me, opp := laviniaBoard(t)

	big := laviniaHandCard(g, opp, "Big Sorcery", "Sorcery", "{2}{G}")
	if err := g.CastSpell(opp.ID, big, game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("mana value 3 with two lands: err = %v, want ErrCantCast", err)
	}
	xSpell := laviniaHandCard(g, opp, "Green X", "Sorcery", "{X}{G}")
	if err := g.CastSpell(opp.ID, xSpell, game.CastSpellParams{Strict: true, XValue: 2}); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("{X}{G} at X=2 with two lands: err = %v, want ErrCantCast", err)
	}
	if !opp.Hand.Contains(big) || !opp.Hand.Contains(xSpell) {
		t.Fatal("a refused cast left the hand")
	}
	if err := g.CastSpell(opp.ID, xSpell, game.CastSpellParams{Strict: true, XValue: 1}); err != nil {
		t.Fatalf("{X}{G} at X=1 with two lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	hydra := laviniaHandCard(g, opp, "Big Hydra", "Creature — Hydra", "{3}{G}")
	if err := g.CastSpell(opp.ID, hydra, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a creature spell above the land count: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hydra) {
		t.Fatal("the creature spell, paid for with mana, did not resolve")
	}

	mine := game.Card{InstanceID: uuid.New(), Name: "My Sorcery", TypeLine: "Sorcery", ManaCost: "{5}{U}", Owner: me.ID, Controller: me.ID}
	if err := effectstest.CastGateFor(g, me.ID, mine, game.ZoneHand); err != nil {
		t.Fatalf("Lavinia restricted her own controller: %v", err)
	}
}

// The bot's move list never offers an opponent an X Lavinia refuses:
// with two lands and mana for far more, {X}{G} is offered at X=1 at most.
func TestLaviniaAzoriusRenegadeBoundsTheEnumeratedX(t *testing.T) {
	g, _, opp := laviniaBoard(t)
	xSpell := laviniaHandCard(g, opp, "Green X", "Sorcery", "{X}{G}")
	offered := false
	for _, m := range legal.EnumerateFor(g, opp.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != xSpell {
			continue
		}
		offered = true
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.XValue > 1 {
			t.Errorf("enumerator offers X=%d, above Lavinia's ceiling of 1", p.XValue)
		}
		if m.Value != nil && m.Value.Max != nil && *m.Value.Max > 1 {
			t.Errorf("enumerator's open X range reaches %d, above Lavinia's ceiling of 1", *m.Value.Max)
		}
	}
	if !offered {
		t.Fatal("the enumerator offers no cast of the X spell at all")
	}
	var ceiling int
	var ok bool
	var card game.Card
	g.ReadSnapshot(func() {
		card, _ = g.LookupCardForEffect(xSpell)
		ceiling, ok = g.CastGateXCeilingForEffect(opp.ID, card, game.ZoneHand)
	})
	if !ok || ceiling != 1 {
		t.Fatalf("cast gate X ceiling = (%d, %v), want (1, true)", ceiling, ok)
	}
}

// An opponent's spell cast for no mana is countered; one paid for with
// mana is not.
func TestLaviniaAzoriusRenegadeCountersOpponentsFreeSpells(t *testing.T) {
	g, _, opp := laviniaBoard(t)
	free := laviniaHandCard(g, opp, "Free Bear", "Creature — Bear", "{0}")
	if err := g.CastSpell(opp.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the free bear: %v", err)
	}
	effectstest.RingSettle(t, g)
	if g.Battlefield.Contains(free) || !opp.Graveyard.Contains(free) {
		t.Fatal("an opponent's spell cast for no mana was not countered")
	}

	paid := laviniaHandCard(g, opp, "Paid Bear", "Creature — Bear", "{1}{G}")
	if err := g.CastSpell(opp.ID, paid, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the paid bear: %v", err)
	}
	effectstest.RingSettle(t, g)
	if !g.Battlefield.Contains(paid) {
		t.Fatal("an opponent's spell paid for with mana was countered")
	}
}
