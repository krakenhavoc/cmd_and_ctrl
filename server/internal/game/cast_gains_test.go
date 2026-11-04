package game

import (
	"encoding/json"
	"strings"
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

// permissionCounterRiderGame is seat 0 in its precombat main holding a
// graveyard cast permission whose casts enter with one `kind` counter
// (#4679), with a creature card and a Kindred instant in its graveyard.
func permissionCounterRiderGame(t *testing.T, kind string) (g *Game, me *Player, source, creature, instant uuid.UUID) {
	t.Helper()
	g = newActiveGame(t)
	me = g.Seats[0]
	toMainPhase(t, g)
	source = uuid.New()
	c := NewCard("Rider Bear", me.ID)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	i := NewCard("Rider Trick", me.ID)
	i.TypeLine = "Kindred Instant — Bear"
	me.Graveyard.PushTop(c)
	me.Graveyard.PushTop(i)
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player: me.ID, Zone: ZoneGraveyard, Scope: ScopeStanding,
			Duration:              IndefiniteDuration(),
			EntersWithCounterKind: kind, EntersWithCounters: 1,
			Source: source, SourceName: "Rider Source",
			Label: "test: cast from your graveyard; that creature enters with a counter",
		})
	})
	return g, me, source, c.InstanceID, i.InstanceID
}

// TestPermissionCounterRiderMarksTheCreatureSpellCastThroughIt: the cast
// writes one PromisedCounters mark naming the permission's source, the
// creature enters with the counter, and a finality counter exiles it
// instead of letting it die (CR 122.1h).
func TestPermissionCounterRiderMarksTheCreatureSpellCastThroughIt(t *testing.T) {
	g, me, source, bear, _ := permissionCounterRiderGame(t, CounterFinality)
	if err := g.CastSpell(me.ID, bear, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast through the permission: %v", err)
	}
	want := PromisedCounter{Kind: CounterFinality, Count: 1, Source: source, SourceName: "Rider Source"}
	if marks := g.StackMeta[bear].PromisedCounters; len(marks) != 1 || marks[0] != want {
		t.Fatalf("marks = %+v, want [%+v]", marks, want)
	}
	passUntilStackEmpty(t, g)
	c := findBattlefieldCard(g, bear)
	if c == nil || c.Counters[CounterFinality] != 1 {
		t.Fatalf("the creature did not enter with one finality counter: %+v", c)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if me.Graveyard.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("a creature with a finality counter that died should be exiled, not in the graveyard")
	}
}

// TestPermissionCounterRiderSkipsANoncreatureSpell: "that creature" — a
// noncreature spell cast through the same permission is not marked.
func TestPermissionCounterRiderSkipsANoncreatureSpell(t *testing.T) {
	g, me, _, _, trick := permissionCounterRiderGame(t, CounterPlusOne)
	if err := g.CastSpell(me.ID, trick, CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast the instant through the permission: %v", err)
	}
	if marks := g.StackMeta[trick].PromisedCounters; len(marks) != 0 {
		t.Fatalf("a noncreature spell was marked: %+v", marks)
	}
}

// TestPermissionCounterRiderRoundTripsThroughJSON: the rider is plain
// data on a stored permission, so a restore point holding it keeps it.
func TestPermissionCounterRiderRoundTripsThroughJSON(t *testing.T) {
	p := CastPermission{EntersWithCounterKind: CounterPlusOne, EntersWithCounters: 1}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var back CastPermission
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.EntersWithCounterKind != CounterPlusOne || back.EntersWithCounters != 1 {
		t.Fatalf("round trip lost the rider: %s", raw)
	}
	// And a permission without the rider writes neither key, so every
	// stored permission already on disk re-encodes as it was.
	raw, err = json.Marshal(CastPermission{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "entersWithCounter") {
		t.Fatalf("an empty rider was written: %s", raw)
	}
}
