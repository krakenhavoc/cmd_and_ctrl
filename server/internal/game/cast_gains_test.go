package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_gains_test.go — #3768: the duration an effect with no stated
// duration about the permanent a spell will become lasts for.

// TestPermanentThisSpellBecomesDurationFollowsTheObject: it holds while
// the spell is on the stack, then on the battlefield under any
// controller, and is over once the object is anywhere else (CR 400.7).
func TestPermanentThisSpellBecomesDurationFollowsTheObject(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	d := ForThePermanentThisSpellBecomesDuration(id)
	holds := func() bool {
		var ok bool
		g.WithWriteLock(func() { ok = g.DurationHoldsForEffect(d) })
		return ok
	}
	if !d.Known() {
		t.Fatalf("the duration is not Known: %s", d.Problem())
	}

	g.Stack.PushTop(Card{InstanceID: id, Name: "Spell Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if !holds() {
		t.Fatal("the duration does not hold while the spell is on the stack")
	}

	c, err := g.Stack.Remove(id)
	if err != nil {
		t.Fatal(err)
	}
	c.Controller = opp.ID
	g.Battlefield.PushTop(c)
	if !holds() {
		t.Fatal("the duration ended on the battlefield under another controller; it has no control test")
	}

	c, err = g.Battlefield.Remove(id)
	if err != nil {
		t.Fatal(err)
	}
	me.Graveyard.PushTop(c)
	if holds() {
		t.Fatal("the duration still holds for a card in a graveyard")
	}
}

// TestPermissionGainsOnlyForAPermanentSpell: an instant cast through a
// permission with gains registers nothing, a creature registers one
// record with both mods.
func TestPermissionGainsOnlyForAPermanentSpell(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	instant, creature := uuid.New(), uuid.New()
	g.Stack.PushTop(Card{InstanceID: instant, Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.Stack.PushTop(Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	gains := permissionCastGainsOf(&CastPermission{GainsKeywords: []string{"haste"}, GainsAbilities: []string{"x/y"}, SourceName: "Test"})
	before := len(g.ScopedEffects)
	g.WithWriteLock(func() {
		g.grantGainsForCastLocked(instant, gains)
	})
	if len(g.ScopedEffects) != before {
		t.Fatal("an instant cast this way registered gains")
	}
	g.WithWriteLock(func() {
		g.grantGainsForCastLocked(creature, gains)
	})
	if len(g.ScopedEffects) != before+1 {
		t.Fatalf("a creature cast this way registered %d records, want 1", len(g.ScopedEffects)-before)
	}
	rec := g.ScopedEffects[len(g.ScopedEffects)-1]
	if len(rec.Mods) != 2 || rec.Duration.Condition != WhileSourceOnceItLands || rec.Affected[0].Controller != uuid.Nil {
		t.Fatalf("record %+v", rec)
	}
}
