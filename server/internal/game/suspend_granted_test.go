package game

import (
	"testing"

	"github.com/google/uuid"
)

// suspend_granted_test.go — "If it doesn't have suspend, it gains suspend"
// (Kang Prime, Doom's Time Platform; #4455). The marker is
// Card.SuspendGranted; the rest is the printed keyword's machinery.

// exileForSuspendGrant puts a card in exile as a player's own and returns its
// id, with no suspend of its own.
func exileForSuspendGrant(g *Game, p *Player, name, typeLine string, power, toughness int) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.Power, c.Toughness = power, toughness
	g.Exile.PushTop(c)
	return c.InstanceID
}

func grantSuspend(t *testing.T, g *Game, p *Player, id uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.GainSuspendForEffect(p.ID, id, n) })
	if err != nil {
		t.Fatalf("GainSuspendForEffect: %v", err)
	}
}

// A card with no suspend that is given counters and suspend counts down on
// its owner's upkeeps, is offered the free cast on the last tick, and the
// creature that results has haste (CR 702.62a).
func TestAGrantedSuspendCountsDownAndCastsFreeWithHaste(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogTriggers(t, func(string) []TriggeredAbility { return nil })
	id := exileForSuspendGrant(g, me, "Test Goblin", "Creature — Goblin", 2, 2)

	grantSuspend(t, g, me, id, 2)
	c := exiledCardByIDLocked(g, id)
	if c == nil || !c.SuspendGranted || c.Counters[CounterTime] != 2 {
		t.Fatalf("after the grant: %+v", c)
	}

	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 1 {
		t.Fatalf("after one upkeep: %d time counters, want 1", n)
	}
	if pendingChoiceOfKind(g, PendingChoiceResolutionCast) != nil {
		t.Fatal("the cast was offered before the last counter came off")
	}

	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 0 {
		t.Fatalf("after two upkeeps: %d time counters, want 0", n)
	}
	if pendingChoiceOfKind(g, PendingChoiceResolutionCast) == nil {
		t.Fatal("the last tick offered no cast")
	}
	perm := grantOn(g, me.ID, id, ZoneExile)
	if perm == nil || perm.Cost != "{0}" || !perm.GrantsHaste {
		t.Fatalf("the offer granted %+v, want a free cast that grants haste", perm)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	settleStack(t, g)
	bf, ok := battlefieldCardByID(g, id)
	if !ok {
		t.Fatal("the creature is not on the battlefield")
	}
	if !HasKeyword(&bf, "haste") {
		t.Error("a creature cast from granted suspend has no haste (CR 702.62a)")
	}
	if bf.SuspendGranted {
		t.Error("the permanent kept the suspend an exiled card was given (CR 400.7)")
	}
}

// Nobody else's upkeep counts it down: the granted triggers are the
// keyword's, owner's upkeep only.
func TestAGrantedSuspendTicksOnlyOnTheOwnersUpkeep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	withCatalogTriggers(t, func(string) []TriggeredAbility { return nil })
	id := exileForSuspendGrant(g, me, "Test Goblin", "Creature — Goblin", 2, 2)
	grantSuspend(t, g, me, id, 3)

	for seat := 1; seat <= 3; seat++ {
		upkeepFor(t, g, seat)
		settleStack(t, g)
	}
	if n := timeCounters(g, id); n != 3 {
		t.Fatalf("after three other upkeeps: %d time counters, want 3", n)
	}
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 2 {
		t.Fatalf("after its owner's upkeep: %d time counters, want 2", n)
	}
}

// A card that prints suspend is not given a second instance: it gets the
// counters, no marker, and one removal per upkeep.
func TestACardThatPrintsSuspendIsNotGivenASecondInstance(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withSuspendCard(t, 3, "{R}")
	advanceTo(t, g, StepPrecombatMain)
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	id := suspendIt(t, g, me, "Rift Bolt", "Sorcery", "{2}{R}")

	grantSuspend(t, g, me, id, 2)
	c := exiledCardByIDLocked(g, id)
	if c == nil || c.SuspendGranted {
		t.Fatalf("a printed-suspend card was marked as given suspend: %+v", c)
	}
	if c.Counters[CounterTime] != 5 {
		t.Fatalf("time counters: %d, want 3 + 2", c.Counters[CounterTime])
	}
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 4 {
		t.Errorf("after one upkeep: %d time counters, want 4 (one countdown trigger)", n)
	}
}

// Giving suspend twice marks the card once, and CardHasSuspend says so.
func TestGrantingSuspendTwiceAddsCountersButOneInstance(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogTriggers(t, func(string) []TriggeredAbility { return nil })
	id := exileForSuspendGrant(g, me, "Test Goblin", "Creature — Goblin", 2, 2)
	if CardHasSuspend(*exiledCardByIDLocked(g, id)) {
		t.Fatal("a plain card reports suspend")
	}
	grantSuspend(t, g, me, id, 2)
	grantSuspend(t, g, me, id, 2)
	c := exiledCardByIDLocked(g, id)
	if !CardHasSuspend(*c) || c.Counters[CounterTime] != 4 {
		t.Fatalf("after two grants: %+v", c)
	}
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if n := timeCounters(g, id); n != 3 {
		t.Errorf("after one upkeep: %d time counters, want 3", n)
	}
}

// CR 400.7: the gain belongs to the exiled object. A card that leaves exile
// and comes back is a new object with no marker and no counters.
func TestGrantedSuspendEndsWhenTheCardLeavesExile(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogTriggers(t, func(string) []TriggeredAbility { return nil })
	id := exileForSuspendGrant(g, me, "Test Goblin", "Creature — Goblin", 2, 2)
	grantSuspend(t, g, me, id, 2)

	if _, err := MoveCard(g.Exile, me.Hand, id); err != nil {
		t.Fatalf("MoveCard: %v", err)
	}
	if c := handCard(me, id); c == nil || c.SuspendGranted {
		t.Fatalf("the card in hand kept the granted suspend: %+v", c)
	}
	if _, err := MoveCard(me.Hand, g.Exile, id); err != nil {
		t.Fatalf("MoveCard back: %v", err)
	}
	c := exiledCardByIDLocked(g, id)
	if c == nil || c.SuspendGranted || CardIsSuspended(*c) {
		t.Fatalf("the card exiled again is suspended: %+v", c)
	}
	upkeepFor(t, g, 0)
	settleStack(t, g)
	if pendingChoiceOfKind(g, PendingChoiceResolutionCast) != nil {
		t.Error("a card that left exile and came back was offered a free cast")
	}
}

// The marker is saved and restored with the card, and the restored game
// still counts it down.
func TestGrantedSuspendSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogTriggers(t, func(string) []TriggeredAbility { return nil })
	id := exileForSuspendGrant(g, me, "Test Goblin", "Creature — Goblin", 2, 2)
	grantSuspend(t, g, me, id, 2)

	_, restored := roundTrip(t, g)
	c := exiledCardByIDLocked(restored, id)
	if c == nil || !c.SuspendGranted || c.Counters[CounterTime] != 2 {
		t.Fatalf("after the round trip: %+v", c)
	}
	upkeepFor(t, restored, 0)
	settleStack(t, restored)
	if n := timeCounters(restored, id); n != 1 {
		t.Errorf("the restored card counted down to %d, want 1", n)
	}
}
