package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// noCostSorcery seeds an Ancestral Vision-shaped card: a sorcery with
// no mana cost, stamped the way deck import stamps it (Layout set,
// ManaCost empty).
func noCostSorcery(me *Player, oracle string) Card {
	c := NewCard("Test Vision", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = ""
	c.Layout = "normal"
	c.OracleID = oracle
	return c
}

// CR 118.6: a spell with no mana cost can't be cast by paying it. The
// refusal holds in every mana mode, because permissive mode and the
// strict-mode override both waive PAYMENT, and there is no cost here
// to pay on paper either.
func TestNoManaCostSpellCantBeCastByPayingIt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params CastSpellParams
	}{
		{"strict", CastSpellParams{Strict: true}},
		{"permissive", CastSpellParams{}},
		{"strict override", CastSpellParams{Strict: true, ForceCast: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			me := g.Seats[0]
			advanceTo(t, g, StepPrecombatMain)
			c := noCostSorcery(me, "test-vision")
			me.Hand.PushTop(c)

			err := g.CastSpell(me.ID, c.InstanceID, tc.params)
			if !errors.Is(err, ErrNoManaCost) {
				t.Fatalf("cast of a no-cost sorcery: got %v, want ErrNoManaCost", err)
			}
			if !me.Hand.Contains(c.InstanceID) {
				t.Errorf("refused cast moved the card out of hand")
			}
			if g.Stack.Contains(c.InstanceID) {
				t.Errorf("refused cast reached the stack")
			}
		})
	}
}

// {0} is a cost, and "no mana cost" is not. Ornithopter casts for
// nothing.
func TestZeroManaCostSpellStillCasts(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Ornithopter", me.ID)
	c.TypeLine = "Artifact Creature — Thopter"
	c.ManaCost = "{0}"
	c.Layout = "normal"
	me.Hand.PushTop(c)

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast of a {0} spell: %v", err)
	}
	if !g.Stack.Contains(c.InstanceID) {
		t.Errorf("{0} spell did not reach the stack")
	}
}

// CR 118.6a: an alternative cost may still be paid on a spell with
// no mana cost.
func TestNoManaCostSpellCastsForAnAlternativeCost(t *testing.T) {
	const oracle = "test-vision"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogAlternativeCosts(t, altCostFor(oracle, AlternativeCost{
		Key: "test-alt", Label: "Cast for {U}", ManaCost: "{U}",
	}))
	advanceTo(t, g, StepPrecombatMain)
	c := noCostSorcery(me, oracle)
	me.Hand.PushTop(c)
	me.ManaPool.AddMana(ManaToken{Color: "U"})

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, AlternativeCost: "test-alt"}); err != nil {
		t.Fatalf("alternative-cost cast of a no-cost sorcery: %v", err)
	}
	if !g.Stack.Contains(c.InstanceID) {
		t.Errorf("alternative-cost cast did not reach the stack")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("alternative cost was not paid: pool %+v", me.ManaPool)
	}
}

// CR 118.6a again: "cast it without paying its mana cost" is an
// alternative cost. Cascade and a Siege's back face reach this
// engine as an exile grant priced at {0}. A grant with no price of
// its own ("you may cast that card", impulse draw) still pays the
// printed cost, so it is refused.
func TestNoManaCostSpellAndExileGrants(t *testing.T) {
	cast := func(t *testing.T, grant func(me uuid.UUID) CastPermission) error {
		t.Helper()
		g := newActiveGame(t)
		me := g.Seats[0]
		advanceTo(t, g, StepPrecombatMain)
		c := noCostSorcery(me, "test-vision")
		g.Exile.PushTop(c)
		g.GrantCastPermissionOverCardForEffect(c.InstanceID, grant(me.ID))
		return g.CastSpell(me.ID, c.InstanceID, CastSpellParams{Strict: true, FromZone: "exile"})
	}

	t.Run("free-cast grant", func(t *testing.T) {
		err := cast(t, func(me uuid.UUID) CastPermission {
			return CastPermission{Player: me, Cost: "{0}"}
		})
		if err != nil {
			t.Fatalf("free-cast grant on a no-cost sorcery: %v", err)
		}
	})
	t.Run("grant without a price", func(t *testing.T) {
		err := cast(t, func(me uuid.UUID) CastPermission {
			return CastPermission{Player: me}
		})
		if !errors.Is(err, ErrNoManaCost) {
			t.Fatalf("unpriced grant on a no-cost sorcery: got %v, want ErrNoManaCost", err)
		}
	})
}

// HasNoManaCost reads only stamped cards. Lands are played, not cast,
// and a fixture with no Layout was never stamped, so its empty cost
// means "unknown" rather than "no mana cost".
func TestHasNoManaCost(t *testing.T) {
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"stamped sorcery, no cost", Card{TypeLine: "Sorcery", Layout: "normal"}, true},
		{"stamped sorcery, {0}", Card{TypeLine: "Sorcery", Layout: "normal", ManaCost: "{0}"}, false},
		{"stamped land", Card{TypeLine: "Basic Land — Island", Layout: "normal"}, false},
		{"unstamped fixture", Card{TypeLine: "Sorcery"}, false},
	} {
		if got := HasNoManaCost(tc.card); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
