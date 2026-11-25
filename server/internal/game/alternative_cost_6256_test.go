package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// alternative_cost_6256_test.go — #6256: "If you control a Forest,
// rather than pay this spell's mana cost, you may have an opponent gain
// 3 life" (Invigorate).
//
//	118.9. Some spells have alternative costs. An alternative cost is a
//	       cost listed in a spell's text, or applied to it from another
//	       effect, that its controller may pay rather than paying the
//	       spell's mana cost. ...
//
// The caster names the opponent, as a gift's recipient and an "an
// opponent gains 5 life" splice cost's are.

// invigorateOffer is Invigorate's cost, with its printed gate.
func invigorateOffer() AlternativeCost {
	return AlternativeCost{
		Key: "invigorate", Label: "Have an opponent gain 3 life rather than pay this spell's mana cost",
		OpponentGainsLife: 3,
		Condition: func(g *Game, controller uuid.UUID) bool {
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.Controller == controller && c.HasSubtype("Forest") {
					return true
				}
			}
			return false
		},
	}
}

// The named opponent gains the life, the caster pays no mana, and a
// claim that names nobody, names the caster, or names a player who is
// not an opponent is refused with nothing paid.
func TestOpponentGainsLifeAlternativeCostNamesAnOpponent(t *testing.T) {
	const oracle = "test-invigorate"
	offer := invigorateOffer()
	g := newFourPlayerActiveGame(t)
	me, them, third := g.Seats[0], g.Seats[1], g.Seats[2]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	forest := NewCard("Forest", me.ID)
	forest.TypeLine, forest.Controller = "Basic Land — Forest", me.ID
	g.Battlefield.PushTop(forest)
	id := blastInHand(t, g, me, oracle)

	var options []uuid.UUID
	var payable bool
	g.WithWriteLock(func() {
		options = g.SpliceOpponentsLocked(me.ID)
		payable = g.AlternativeCostPayableLocked(me.ID, id, &offer)
	})
	if !payable {
		t.Fatal("the offer is withheld although a Forest is out and opponents can gain life")
	}
	if len(options) != 3 || containsID(options, me.ID) {
		t.Fatalf("opponents = %v, want the caster's three opponents", options)
	}

	for name, params := range map[string]CastSpellParams{
		"no opponent named": {Strict: true, AlternativeCost: "invigorate"},
		"the caster":        {Strict: true, AlternativeCost: "invigorate", AltCostOpponent: me.ID},
		"nobody at all":     {Strict: true, AlternativeCost: "invigorate", AltCostOpponent: uuid.New()},
	} {
		if err := g.CastSpell(me.ID, id, params); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: %v, want ErrInvalidParam", name, err)
		}
	}
	life, thirdLife := them.Life, third.Life
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "invigorate", AltCostOpponent: them.ID}); err != nil {
		t.Fatalf("cast for the alternative cost: %v", err)
	}
	if them.Life != life+3 {
		t.Errorf("the named opponent's life %d → %d, want +3", life, them.Life)
	}
	if third.Life != thirdLife {
		t.Errorf("an opponent who was not named gained life: %d → %d", thirdLife, third.Life)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %v, want nothing spent", me.ManaPool)
	}
}

// Without a Forest the offer is not available, and the claim is refused.
// An opponent sent on a cast that claims no such offer is refused too.
func TestOpponentGainsLifeAlternativeCostGateAndStrayOpponent(t *testing.T) {
	const oracle = "test-invigorate-gate"
	offer := invigorateOffer()
	g := newFourPlayerActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	id := blastInHand(t, g, me, oracle)

	var payable bool
	g.WithWriteLock(func() { payable = g.AlternativeCostPayableLocked(me.ID, id, &offer) })
	if payable {
		t.Error("the offer is available with no Forest out")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "invigorate", AltCostOpponent: them.ID}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("claiming it with no Forest: %v, want ErrInvalidParam", err)
	}
	if them.Life != 40 {
		t.Errorf("a refused cast gave life: %d", them.Life)
	}

	// A stray opponent on a printed-cost cast is refused, not ignored.
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AltCostOpponent: them.ID}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("a stray alt_cost_opponent: %v, want ErrInvalidParam", err)
	}
}

// CR 614.17b: an opponent who can't gain life can't be given it, so a
// table where no opponent can gain is not offered the cost at all.
func TestOpponentGainsLifeAlternativeCostNeedsSomebodyWhoCanGain(t *testing.T) {
	const oracle = "test-invigorate-cant-gain"
	offer := invigorateOffer()
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	withCatalogAlternativeCosts(t, altCostFor(oracle, offer))
	forest := NewCard("Forest", me.ID)
	forest.TypeLine, forest.Controller = "Basic Land — Forest", me.ID
	g.Battlefield.PushTop(forest)
	id := blastInHand(t, g, me, oracle)
	for _, p := range g.Seats {
		if p.ID != me.ID {
			p.Eliminated = true
		}
	}
	var payable bool
	g.WithWriteLock(func() { payable = g.AlternativeCostPayableLocked(me.ID, id, &offer) })
	if payable {
		t.Error("the offer is available with no opponent left to gain the life (CR 614.17b)")
	}
}
