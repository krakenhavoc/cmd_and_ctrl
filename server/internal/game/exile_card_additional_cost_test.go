package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exile_card_additional_cost_test.go — #3536: "As an additional cost to
// cast this spell, exile N cards from your graveyard" (Abhorrent Oculus) on
// a spell's mandatory additional-cost slot. The catalog is stubbed so the
// engine path is tested without a card: the spell is a three-card version.

const exileFuelOracle = "test-exile-fuel"

func withExileFuelCost(t *testing.T, n int) {
	t.Helper()
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id != exileFuelOracle {
			return nil
		}
		return &AdditionalCost{
			ExileCards: &ExileCost{N: n, Label: "cards", From: ZoneGraveyard},
			Label:      "Exile cards from your graveyard",
		}
	})
}

func exileFuelSpell(owner uuid.UUID) Card {
	c := NewCard("Test Eye", owner)
	c.TypeLine = "Creature — Eye"
	c.ManaCost = "{2}{U}"
	c.Layout = "normal"
	c.OracleID = exileFuelOracle
	return c
}

func graveyardFuel(p *Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		c := NewCard("Fuel", p.ID)
		c.TypeLine = "Instant"
		ids[i] = c.InstanceID
		p.Graveyard.PushTop(c)
	}
	return ids
}

// castExileFuelSpell puts the spell in hand, walks to a main phase and
// casts it with `params` (strict mana off: the cost under test is the
// exile, not the mana).
func castExileFuelSpell(t *testing.T, g *Game, me *Player, params CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := exileFuelSpell(me.ID)
	me.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(me.ID, c.InstanceID, params)
}

func TestExileCardsCostExilesWithTheSpellOnTheStackAndRecordsThem(t *testing.T) {
	withExileFuelCost(t, 3)
	g := newActiveGame(t)
	me := g.Seats[0]
	fuel := graveyardFuel(me, 4)

	spell, err := castExileFuelSpell(t, g, me, CastSpellParams{ExileIDs: fuel[:3]})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell is not on the stack")
	}
	for _, id := range fuel[:3] {
		if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
			t.Errorf("fuel %s was not exiled", id)
		}
	}
	if !me.Graveyard.Contains(fuel[3]) {
		t.Error("the fourth card, which was not named, left the graveyard")
	}
	paid := g.StackMeta[spell].Paid
	if len(paid.Exiled) != 3 {
		t.Fatalf("Paid.Exiled = %v, want the three named cards", paid.Exiled)
	}
	for i, id := range fuel[:3] {
		if paid.Exiled[i] != id {
			t.Errorf("Paid.Exiled[%d] = %s, want %s", i, paid.Exiled[i], id)
		}
	}
}

// CR 118.3: a graveyard that cannot cover the count refuses the cast and
// nothing moves, whether the player names too few or the cards don't exist.
func TestExileCardsCostIsRefusedWithTooFewCards(t *testing.T) {
	withExileFuelCost(t, 3)
	g := newActiveGame(t)
	me := g.Seats[0]
	fuel := graveyardFuel(me, 2)

	for name, ids := range map[string][]uuid.UUID{
		"two named of three":  fuel,
		"nothing named":       nil,
		"four named of three": append(append([]uuid.UUID(nil), fuel...), uuid.New(), uuid.New()),
	} {
		t.Run(name, func(t *testing.T) {
			spell, err := castExileFuelSpell(t, g, me, CastSpellParams{ExileIDs: ids})
			if !errors.Is(err, ErrInvalidParam) && !errors.Is(err, ErrCardNotFound) {
				t.Fatalf("cast = %v, want a refusal of the payment", err)
			}
			if g.Stack.Contains(spell) {
				t.Error("a refused cast reached the stack")
			}
			if me.Graveyard.Size() != 2 || g.Exile.Size() != 0 {
				t.Errorf("a refused cast moved cards: graveyard %d, exile %d", me.Graveyard.Size(), g.Exile.Size())
			}
		})
	}
}

// A card named twice, one that is not in the caster's graveyard (an
// opponent's, one in hand, the spell itself), or one also named to delve
// is refused, and nothing moves.
func TestExileCardsCostRefusesAnIllegalPick(t *testing.T) {
	withExileFuelCost(t, 2)
	prevDelve := CatalogDelve
	CatalogDelve = func(key string) bool { return key == exileFuelOracle }
	t.Cleanup(func() { CatalogDelve = prevDelve })

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := graveyardFuel(me, 3)
	theirs := graveyardFuel(opp, 1)
	inHand := NewCard("In Hand", me.ID)
	me.Hand.PushTop(inHand)

	for _, tc := range []struct {
		name   string
		params CastSpellParams
	}{
		{"the same card twice", CastSpellParams{ExileIDs: []uuid.UUID{mine[0], mine[0]}}},
		{"an opponent's graveyard card", CastSpellParams{ExileIDs: []uuid.UUID{mine[0], theirs[0]}}},
		{"a card in hand", CastSpellParams{ExileIDs: []uuid.UUID{mine[0], inHand.InstanceID}}},
		{"a card also named to delve", CastSpellParams{ExileIDs: mine[:2], DelveIDs: mine[1:2]}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spell, err := castExileFuelSpell(t, g, me, tc.params)
			if err == nil {
				t.Fatal("the cast was accepted")
			}
			if g.Stack.Contains(spell) {
				t.Error("a refused cast reached the stack")
			}
			if me.Graveyard.Size() != 3 || opp.Graveyard.Size() != 1 || g.Exile.Size() != 0 {
				t.Errorf("a refused cast moved cards: graveyard %d/%d, exile %d", me.Graveyard.Size(), opp.Graveyard.Size(), g.Exile.Size())
			}
		})
	}
}

// exile_ids on a spell with no such cost is refused rather than ignored,
// as an unexpected sacrifice_ids is.
func TestExileIDsWithoutAnExileCardsCostAreRefused(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fuel := graveyardFuel(me, 1)
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Plain Spell", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{U}"
	c.Layout = "normal"
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{ExileIDs: fuel}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("cast = %v, want ErrInvalidParam", err)
	}
	if !me.Graveyard.Contains(fuel[0]) {
		t.Error("a refused cast exiled the card")
	}
}

// The Oculus ruling (2024-09-20): the additional cost must be paid even if
// the spell is cast "without paying its mana cost". The exile grant is a
// {0} cast; the six still have to come out of the graveyard.
func TestExileCardsCostIsDemandedUnderAFreeCast(t *testing.T) {
	withExileFuelCost(t, 3)
	g := newActiveGame(t)
	me := g.Seats[0]
	fuel := graveyardFuel(me, 3)
	id := exileWithGrant(t, g, exileFuelSpell(me.ID), freeCastGrant(me.ID))

	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile"})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("a free cast naming no cards = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(id) || me.Graveyard.Size() != 3 {
		t.Fatal("the refused free cast moved something")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile", ExileIDs: fuel}); err != nil {
		t.Fatalf("the free cast naming the three: %v", err)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard = %d cards, want all three exiled", me.Graveyard.Size())
	}
	// Three graveyard cards plus the spell's own card leaving exile for
	// the stack: the three paid cards are in exile now.
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Errorf("fuel %s is not in exile", id)
		}
	}
}

// The spell is never a candidate for its own cost: a card cast from its
// own graveyard is excluded from the options the view stamps and the
// validator accepts.
func TestExileCardsCostNeverOffersTheSpellItself(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fuel := graveyardFuel(me, 2)
	self := exileFuelSpell(me.ID)
	me.Graveyard.PushTop(self)
	cost := &ExileCost{N: 2, Label: "cards", From: ZoneGraveyard}

	var options []uuid.UUID
	g.ReadSnapshot(func() { options = g.ExileCostOptionsForEffect(me.ID, self.InstanceID, cost) })
	if len(options) != 2 {
		t.Fatalf("options = %v, want only the two other cards", options)
	}
	for _, id := range options {
		if id == self.InstanceID {
			t.Fatal("the spell offers itself as its own fuel")
		}
	}
	var err error
	g.WithWriteLock(func() {
		_, err = g.validateExileCardsCostLocked(me.ID, self.InstanceID, cost, []uuid.UUID{fuel[0], self.InstanceID}, nil)
	})
	if !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming the spell as fuel = %v, want ErrInvalidParam", err)
	}
}
