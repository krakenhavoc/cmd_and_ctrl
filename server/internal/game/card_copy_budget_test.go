package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// card_copy_budget_test.go — CastPermission.ManaValueBudget (#2017, ADR
// 0066's 2026-10-14 amendment): "any number of spells with total mana
// value N or less from among the copies". Engine-side, uncatalogued
// fixtures: only the shared running total is under test.

func budgetOffer(p *Player, budget int) CardCopyOffer {
	o := freeOffer(p)
	o.ManaValueBudget = &budget
	return o
}

func budgetPermissions(p *Player) []CastPermission {
	var out []CastPermission
	for _, perm := range p.CastPermissions {
		if perm.ManaValueBudget != nil {
			out = append(out, perm)
		}
	}
	return out
}

// The copies are ONE permission naming every copy, and casting spends its
// mana value from the total, shrinking the named set.
func TestManaValueBudgetIsOnePermissionSpentByEachCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := copyFixtureInGraveyard(me, "Fixture Three", "Sorcery", "{2}{R}")
	b := copyFixtureInGraveyard(me, "Fixture Two", "Instant", "{1}{R}")
	c := copyFixtureInGraveyard(me, "Fixture Two B", "Instant", "{1}{B}")

	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(budgetOffer(me, 6), []uuid.UUID{a, b, c})
	})
	if len(copies) != 3 {
		t.Fatalf("copies = %d, want 3", len(copies))
	}
	perms := budgetPermissions(me)
	if len(perms) != 1 || len(perms[0].Cards) != 3 || *perms[0].ManaValueBudget != 6 {
		t.Fatalf("permissions = %+v, want one over three copies with a budget of 6", perms)
	}

	cast := func(id uuid.UUID) error {
		return g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", Strict: true})
	}
	if err := cast(copies[0]); err != nil {
		t.Fatalf("cast the mana value 3 copy: %v", err)
	}
	perms = budgetPermissions(me)
	if len(perms) != 1 || len(perms[0].Cards) != 2 || *perms[0].ManaValueBudget != 3 {
		t.Fatalf("after one cast: %+v, want two copies named and 3 left", perms)
	}
	if err := cast(copies[1]); err != nil {
		t.Fatalf("cast a mana value 2 copy (total 5): %v", err)
	}
	if err := cast(copies[2]); !errors.Is(err, ErrSpellManaValueTooHigh) {
		t.Fatalf("a mana value 2 copy with 1 left: err = %v, want ErrSpellManaValueTooHigh", err)
	}
}

// The remaining total is a fresh value, never written in place: a shallow
// copy of the permissions taken before a cast keeps the total it had.
func TestManaValueBudgetIsNeverSpentInPlace(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := copyFixtureInGraveyard(me, "Fixture Three", "Sorcery", "{2}{R}")
	b := copyFixtureInGraveyard(me, "Fixture Two", "Instant", "{1}{R}")
	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(budgetOffer(me, 6), []uuid.UUID{a, b})
	})
	before := append([]CastPermission(nil), me.CastPermissions...)
	if err := g.CastSpell(me.ID, copies[0], CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatal(err)
	}
	if got := *before[0].ManaValueBudget; got != 6 {
		t.Errorf("the earlier permission value total changed to %d under a later cast", got)
	}
	if got := *budgetPermissions(me)[0].ManaValueBudget; got != 3 {
		t.Errorf("the live permission's total is %d, want 3", got)
	}
}

// No budget keeps the earlier shape: one offer per copy and no total.
func TestNoManaValueBudgetKeepsOneOfferPerCopy(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	a := copyFixtureInGraveyard(me, "Fixture Three", "Sorcery", "{2}{R}")
	b := copyFixtureInGraveyard(me, "Fixture Two", "Instant", "{1}{R}")
	g.WithWriteLock(func() {
		g.CopyCardsAndOfferCastsForEffect(freeOffer(me), []uuid.UUID{a, b})
	})
	if n := len(me.CastPermissions); n != 2 {
		t.Errorf("%d permissions, want one per copy", n)
	}
	if n := len(budgetPermissions(me)); n != 0 {
		t.Errorf("%d budgeted permissions with no budget asked for", n)
	}
}
