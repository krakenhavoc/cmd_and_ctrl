package game

import (
	"testing"
)

// intrinsic_land_rows_test.go pins #2113's first half: a land with one
// basic land type gets the shared, read-only CR 305.6 row instead of a
// fresh slice per call.

// withoutCacheChecks turns the test binary's hit-and-verify off for the
// allocation counts: the check rebuilds the baseline on every read,
// which is exactly the allocation being measured around.
func withoutCacheChecks(t *testing.T) {
	t.Helper()
	prev := printedCacheChecks
	printedCacheChecks = false
	t.Cleanup(func() { printedCacheChecks = prev })
}

// basicLandCard returns a basic land with its printed-characteristic
// cache stamped, as a card that has entered a zone has it: the view
// never sees a cold card, and a cold one would pay for a type-line parse.
func basicLandCard(sub string) Card {
	c := Card{Name: sub, TypeLine: "Basic Land — " + sub}
	c.stampPrinted()
	return c
}

func TestOneTypeLandIntrinsicRowIsSharedAndAllocationFree(t *testing.T) {
	withoutCacheChecks(t)
	for i, lt := range landTypeMana {
		c := basicLandCard(lt.Subtype)
		rows := intrinsicLandManaAbilities(&c)
		if len(rows) != 1 || rows[0].Produced != lt.Produced || rows[0].Label != lt.Label || !rows[0].TapCost {
			t.Fatalf("%s: rows = %+v, want one {T}: %s", lt.Subtype, rows, lt.Label)
		}
		if &rows[0] != &intrinsicLandRows[i][0] {
			t.Errorf("%s: the row is not the shared one", lt.Subtype)
		}
		if cap(rows) != len(rows) {
			t.Errorf("%s: cap %d, len %d; want the capacity clipped so an append cannot write into the shared array", lt.Subtype, cap(rows), len(rows))
		}
		if got := testing.AllocsPerRun(50, func() { _ = intrinsicLandManaAbilities(&c) }); got != 0 {
			t.Errorf("%s: intrinsicLandManaAbilities allocated %.0f times; want 0", lt.Subtype, got)
		}
	}
}

// The origin slice rides the same fast path: no declared half and no
// grant means the shared origin row, and neither spelling of the read
// allocates.
func TestOneTypeLandManaRowsAndOriginsAreAllocationFree(t *testing.T) {
	withoutCacheChecks(t)
	c := basicLandCard("Forest")
	rows, origins := ManaAbilitiesWithOriginsOf(&c)
	if len(rows) != 1 || len(origins) != 1 {
		t.Fatalf("rows %d, origins %d; want one of each", len(rows), len(origins))
	}
	if want := IntrinsicLandAbilityRef("G"); origins[0].Ref != want || origins[0].Granted() {
		t.Errorf("origin = %+v, want the intrinsic ref %q", origins[0], want)
	}
	if got := origins.Ref(0); got != IntrinsicLandAbilityRef("G") {
		t.Errorf("origins.Ref(0) = %q", got)
	}
	if got := testing.AllocsPerRun(50, func() { _, _ = ManaAbilitiesWithOriginsOf(&c) }); got != 0 {
		t.Errorf("ManaAbilitiesWithOriginsOf allocated %.0f times; want 0", got)
	}
	if got := testing.AllocsPerRun(50, func() { _ = manaAbilitiesOf(&c) }); got != 0 {
		t.Errorf("manaAbilitiesOf allocated %.0f times; want 0", got)
	}
}

// A land with two basic land types still builds its own two rows, in
// subtype order, and an append to a shared row cannot reach the shared
// array.
func TestTwoTypeLandAndAppendsLeaveTheSharedRowsAlone(t *testing.T) {
	c := Card{Name: "Sandsteppe Citadel", TypeLine: "Land — Plains Forest"}
	rows := intrinsicLandManaAbilities(&c)
	if len(rows) != 2 || rows[0].Produced != "{W}" || rows[1].Produced != "{G}" {
		t.Fatalf("rows = %+v, want {W} then {G}", rows)
	}
	for i := range landTypeMana {
		if &rows[0] == &intrinsicLandRows[i][0] {
			t.Errorf("a two-type land's rows alias the shared row %d", i)
		}
	}

	f := basicLandCard("Forest")
	shared := intrinsicLandManaAbilities(&f)
	grown := append(shared, ManaAbilityShape{Produced: "{C}"})
	grown[0].Produced = "{X}"
	again := intrinsicLandManaAbilities(&f)
	if len(again) != 1 || again[0].Produced != "{G}" {
		t.Errorf("after an append the shared row reads %+v, want one {G}", again)
	}
}

// Subtype matching is unchanged: case-insensitive, duplicates collapse
// to one row, and a land with no basic land type has none.
func TestIntrinsicLandRowsKeepTheirMatching(t *testing.T) {
	for line, want := range map[string]int{
		"Basic Land — Forest":         1,
		"Land — forest":               1,
		"Land — Forest Forest":        1,
		"Land — Forest Island Swamp":  3,
		"Land — Desert":               0,
		"Land":                        0,
		"Creature — Forest Dryad":     0,
		"Land — Plains Island Plains": 2,
	} {
		c := Card{Name: line, TypeLine: line}
		if got := len(intrinsicLandManaAbilities(&c)); got != want {
			t.Errorf("%q: %d rows, want %d", line, got, want)
		}
	}
	for _, line := range []string{"Land — Forest Dryad", "Artifact Land — Forest"} {
		c := Card{Name: line, TypeLine: line}
		if got := len(intrinsicLandManaAbilities(&c)); got != 1 {
			t.Errorf("%q: %d rows, want 1", line, got)
		}
	}
}
