package game

import (
	"testing"

	"github.com/google/uuid"
)

// scoped_effects_entry_key_test.go pins #1558 item 5: an UNSTAMPED
// permanent — one put onto the battlefield without the zone-move event
// that stamps EnteredBattlefieldAt — is pinned exactly, not as the
// EnteredAt 0 wildcard. The wildcard followed it through a flicker, so
// the returned object, which CR 400.7 says is a new one, kept the
// effect. The legacy closure path compared `== stamp` and never did.

// pushUnstampedCreature seeds a Bear the way a fixture does: straight
// onto the battlefield, no zone-move event, EnteredBattlefieldAt 0.
func pushUnstampedCreature(g *Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// reenterForTest is the half of a flicker the object model sees: the
// same instance comes back through a zone move, and the listener
// stamps it as a new object.
func reenterForTest(g *Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventZoneMove, CardID: id, OldZone: ZoneExile, NewZone: ZoneBattlefield})
	})
}

func TestAFlickeredUnstampedPermanentIsNoLongerAffected(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushUnstampedCreature(g, me)

	var d Duration
	var affected []AffectedObject
	g.WithWriteLock(func() {
		d = g.PinnedTo(IndefiniteDuration(), bear)
		affected = g.PinnedObjectsLocked(bear)
	})
	if len(affected) != 1 || !affected[0].Unstamped {
		t.Fatalf("pinned %+v, want the unstamped Bear pinned Unstamped, not the 0 wildcard", affected)
	}
	if !d.PinnedUnstamped {
		t.Fatalf("duration pin %+v, want PinnedUnstamped", d)
	}
	registerScopedEffectForTest(t, g, bear, []Mod{AddSubtypesMod("Orc")}, d)
	if c := scopedEffectChar(t, g, bear); !typeListHas(c.Subtypes, "Orc") {
		t.Fatalf("before the flicker: subtypes %v, want Orc", c.Subtypes)
	}

	reenterForTest(g, bear)
	var stamp int64
	g.WithWriteLock(func() {
		if c, ok := g.battlefieldCardLocked(bear); ok {
			stamp = c.EnteredBattlefieldAt
		}
	})
	if stamp == 0 {
		t.Fatal("the re-entry did not stamp the Bear; the test is not flickering anything")
	}
	if c := scopedEffectChar(t, g, bear); typeListHas(c.Subtypes, "Orc") {
		t.Errorf("after the flicker: subtypes %v — the new object kept an effect pinned to the old one (CR 400.7)", c.Subtypes)
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("%d scoped effects after the flicker, want 0 — the pin names an object that is gone", n)
	}
}

// TestTheEntryWildcardStillFollowsItsObject: EnteredAt 0 is still the
// wildcard for the one record that writes it on purpose — suspend's
// haste, which names a card on the stack before it has a stamp to pin
// (CR 702.62a) — and for every record written before #1558.
func TestTheEntryWildcardStillFollowsItsObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0].ID
	bear := pushUnstampedCreature(g, me)
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, []AffectedObject{{ID: bear}},
			[]Mod{AddKeywordsMod("haste")}, IndefiniteDuration(), "wildcard")
	})
	reenterForTest(g, bear)
	if c := scopedEffectChar(t, g, bear); !hasAbility(c.Abilities, "haste") {
		t.Errorf("abilities %v: the 0 wildcard stopped following its instance", c.Abilities)
	}
}

// TestAnUnstampedPinRoundTrips: the pin survives capture and restore,
// so a restored board still ends the effect on a flicker.
func TestAnUnstampedPinRoundTrips(t *testing.T) {
	g := newActiveGame(t)
	bear := pushUnstampedCreature(g, g.Seats[0].ID)
	var d Duration
	g.WithWriteLock(func() { d = g.PinnedTo(IndefiniteDuration(), bear) })
	registerScopedEffectForTest(t, g, bear, []Mod{AddSubtypesMod("Orc")}, d)

	restored, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if n := len(restored.ScopedEffects); n != 1 ||
		!restored.ScopedEffects[0].Affected[0].Unstamped || !restored.ScopedEffects[0].Duration.PinnedUnstamped {
		t.Fatalf("restored records %+v, want the unstamped pin carried", restored.ScopedEffects)
	}
	reenterForTest(restored, bear)
	if c := scopedEffectChar(t, restored, bear); typeListHas(c.Subtypes, "Orc") {
		t.Errorf("restored board, after the flicker: subtypes %v, want no Orc", c.Subtypes)
	}
}

func TestEntryMatches(t *testing.T) {
	for _, tc := range []struct {
		enteredAt int64
		unstamped bool
		stamp     int64
		want      bool
	}{
		{0, false, 0, true}, {0, false, 42, true}, // the wildcard
		{0, true, 0, true}, {0, true, 42, false}, // pinned while unstamped
		{42, false, 42, true}, {42, false, 0, false}, {42, false, 43, false},
	} {
		if got := entryMatches(tc.enteredAt, tc.unstamped, tc.stamp); got != tc.want {
			t.Errorf("entryMatches(%d, %v, %d) = %v, want %v", tc.enteredAt, tc.unstamped, tc.stamp, got, tc.want)
		}
	}
	id := uuid.New()
	if p := PinObject(id, 0); !p.Unstamped || p.EnteredAt != 0 {
		t.Errorf("PinObject(unstamped) = %+v", p)
	}
	if p := PinObject(id, 7); p.Unstamped || p.EnteredAt != 7 {
		t.Errorf("PinObject(7) = %+v", p)
	}
}

// TestAnUnstampedPinIsAKnownField: the new member field is covered by
// P4's unknown-field refusal, which is what makes it additive within
// v7 — an older binary refuses a file carrying one rather than reading
// the member as a wildcard.
func TestAnUnstampedPinIsAKnownField(t *testing.T) {
	if !affectedObjectJSONKeys["unstamped"] {
		t.Errorf("affectedObjectJSONKeys = %v, want unstamped", affectedObjectJSONKeys)
	}
	if !durationJSONKeys["PinnedUnstamped"] {
		t.Errorf("durationJSONKeys = %v, want PinnedUnstamped", durationJSONKeys)
	}
}

func hasAbility(list []string, want string) bool {
	for _, a := range list {
		if a == want {
			return true
		}
	}
	return false
}
