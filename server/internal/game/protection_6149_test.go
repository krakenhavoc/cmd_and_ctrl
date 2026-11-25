package game

import (
	"testing"

	"github.com/google/uuid"
)

// protection_6149_test.go pins #6149's two qualities (CR 702.16a): an
// exact mana value, and "permanents with <kind> counters on them", read
// off the source's counters as the check is made.

func TestProtectionFromManaValueIsExact(t *testing.T) {
	q, ok := ParseProtectionQuality(ProtectionFromManaValue(3))
	if !ok || q.Kind != ProtectionQualityManaValue || q.Value != "3" || q.Kind.String() != "mana_value" {
		t.Fatalf("parsed %+v, %v", q, ok)
	}
	for mv, want := range map[int]bool{2: false, 3: true, 4: false} {
		src := &Characteristic{SourceManaValue: mv, SourceManaValueKnown: true}
		if got := q.Matches(src); got != want {
			t.Errorf("mana value %d: matches = %v, want %v", mv, got, want)
		}
	}
	if q.Matches(&Characteristic{SourceManaValue: 3}) {
		t.Error("an unpriced source matched protection from mana value 3")
	}
	if ProtectionFromManaValue(-1) != "" {
		t.Error("a negative mana value minted a token")
	}
	// The bounded forms still parse as they did.
	if q, _ := ParseProtectionQuality("protection from mana value 3 or less"); q.Kind != ProtectionQualityManaValueAtMost {
		t.Errorf("'mana value 3 or less' parsed as %v", q.Kind)
	}
}

func TestProtectionFromPermanentsWithCounters(t *testing.T) {
	q, ok := ParseProtectionQuality("protection from permanents with corruption counters on them")
	if !ok || q.Kind != ProtectionQualityPermanentsWithCounter || q.Value != "corruption" || q.Kind.String() != "permanents_with_counter" {
		t.Fatalf("parsed %+v, %v", q, ok)
	}
	if _, ok := ParseProtectionQuality("protection from permanents with  counters on them"); ok {
		t.Error("a quality naming no counter kind parsed")
	}
	if q.Matches(&Characteristic{SourceIsPermanent: true}) {
		t.Error("a permanent with no counters matched")
	}
	if !q.Matches(&Characteristic{SourceIsPermanent: true, SourceCounters: map[string]int{"corruption": 1}}) {
		t.Error("a permanent with a corruption counter did not match")
	}
	if q.Matches(&Characteristic{SourceCounters: map[string]int{"corruption": 1}}) {
		t.Error("a non-permanent source matched")
	}

	// Through SourceCharacteristics on a real battlefield permanent, and
	// through its last-known record once it has left.
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	src := pushTypedTestCard(g, Card{Name: "Ogre", TypeLine: "Creature — Ogre", Owner: seat, Controller: seat,
		Counters: map[string]int{"corruption": 1}})
	c := layeredBattlefieldCard(t, g, src)
	if !q.Matches(SourceCharacteristics(&c)) {
		t.Error("SourceCharacteristics dropped the counters of a permanent source")
	}
	rec := PermanentInfo{Counters: map[string]int{"corruption": 2}}
	if !q.Matches(lastKnownSourceCharacteristics(rec)) {
		t.Error("a departed source's last-known counters were dropped")
	}
	_ = uuid.Nil
}

// Damage from a permanent with a corruption counter is prevented (CR
// 702.16e), through the damage path's own source snapshot.
func TestProtectionFromCounteredPermanentsPreventsDamage(t *testing.T) {
	g := newActiveGame(t)
	seat := g.Seats[0].ID
	walker := pushTypedTestCard(g, Card{Name: "Geyadrone", TypeLine: "Legendary Planeswalker — Dihada", Owner: seat, Controller: seat,
		Keywords: []string{"protection from permanents with corruption counters on them"}, Counters: map[string]int{CounterLoyalty: 4}})
	ogre := pushTypedTestCard(g, Card{Name: "Ogre", TypeLine: "Creature — Ogre", Owner: seat, Controller: seat})
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ogre, "corruption", 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(ogre, walker, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := findBattlefieldCard(g, walker).Counters[CounterLoyalty]; got != 4 {
		t.Errorf("loyalty after damage from a corrupted Ogre = %d, want 4", got)
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ogre, "corruption", -1); err != nil {
			t.Fatal(err)
		}
		if err := g.DealDamageToCreatureForEffect(ogre, walker, 2); err != nil {
			t.Fatal(err)
		}
	})
	if got := findBattlefieldCard(g, walker).Counters[CounterLoyalty]; got != 2 {
		t.Errorf("loyalty after damage from a clean Ogre = %d, want 2", got)
	}
}
