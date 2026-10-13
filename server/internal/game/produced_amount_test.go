package game

import (
	"runtime"
	"testing"
)

// produced_amount_test.go — #3042: a count-driven mana ability hands the
// engine a ManaAmount, and the slots come straight from the count.

func TestSlotsFromAmountShapes(t *testing.T) {
	green := []string{"G"}
	cases := []struct {
		name      string
		in        ManaAmount
		wantSlots int
		wantCap   bool
		oneColor  bool // a single slot carrying per-colour amounts
	}{
		{"single colour", ManaAmount{Options: green, N: 3}, 3, false, false},
		{"zero adds nothing", ManaAmount{Options: green, N: 0}, 0, false, false},
		{"negative adds nothing", ManaAmount{Options: green, N: -4}, 0, false, false},
		{"no options adds nothing", ManaAmount{N: 3}, 0, false, false},
		{"any combination is N independent picks", ManaAmount{Options: AllColors, N: 4}, 4, false, false},
		{"any one colour is one slot", ManaAmount{Options: AllColors, N: 4, OneColor: true}, 1, false, true},
		{"one colour of one is a plain pick", ManaAmount{Options: AllColors, N: 1, OneColor: true}, 1, false, false},
		{"exactly the ceiling is not a cut", ManaAmount{Options: green, N: MaxProducedMana}, MaxProducedMana, false, false},
		{"past the ceiling is cut and announced", ManaAmount{Options: green, N: MaxProducedMana + 1}, MaxProducedMana, true, false},
		{"2^40 is cut", ManaAmount{Options: AllColors, N: 1 << 40}, MaxProducedMana, true, false},
		{"2^40 of one colour is cut", ManaAmount{Options: AllColors, N: 1 << 40, OneColor: true}, 1, true, true},
	}
	for _, tc := range cases {
		slots, capped := slotsFromAmount(tc.in)
		if len(slots) != tc.wantSlots || capped != tc.wantCap {
			t.Errorf("%s: %d slots capped=%v, want %d capped=%v", tc.name, len(slots), capped, tc.wantSlots, tc.wantCap)
			continue
		}
		if tc.oneColor {
			for _, c := range tc.in.Options {
				want := tc.in.N
				if want > MaxProducedMana {
					want = MaxProducedMana
				}
				if got := slots[0].AmountFor(c); got != want {
					t.Errorf("%s: %s amount = %d, want %d", tc.name, c, got, want)
				}
			}
		}
	}
}

// The structured path must produce what parsing the equivalent text
// produced, slot for slot — it is the same declaration, spelled as data.
func TestSlotsFromAmountMatchesTheParsedText(t *testing.T) {
	pairs := []struct {
		in   ManaAmount
		text string
	}{
		{ManaAmount{Options: []string{"B"}, N: 3}, "{B}{B}{B}"},
		{ManaAmount{Options: []string{"U", "R"}, N: 2}, "{U|R}{U|R}"},
		{ManaAmount{Options: AllColors, N: 3, OneColor: true}, "{W3|U3|B3|R3|G3}"},
	}
	for _, p := range pairs {
		got, _ := slotsFromAmount(p.in)
		want, err := ParseProducedMana(p.text)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) {
			t.Fatalf("%q: %d slots, want %d", p.text, len(got), len(want))
		}
		for i := range got {
			if !equalStrings(got[i].Options, want[i].Options) {
				t.Errorf("%q slot %d options %v, want %v", p.text, i, got[i].Options, want[i].Options)
			}
			for _, c := range want[i].Options {
				if got[i].AmountFor(c) != want[i].AmountFor(c) {
					t.Errorf("%q slot %d: %s amount %d, want %d", p.text, i, c, got[i].AmountFor(c), want[i].AmountFor(c))
				}
			}
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// A count of five million is cut before anything is built: no text at
// all, and the slots allocated are the ceiling's, not the request's.
func TestSlotsFromAmountDoesNotAllocateTheAskedForAmount(t *testing.T) {
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	slots, capped := slotsFromAmount(ManaAmount{Options: AllColors, N: 5_000_000})
	runtime.ReadMemStats(&after)
	if !capped || len(slots) != MaxProducedMana {
		t.Fatalf("len=%d capped=%v, want %d slots, capped", len(slots), capped, MaxProducedMana)
	}
	// 1024 entries of a few words each; a text of 5M slots would be tens
	// of megabytes.
	if grew := after.TotalAlloc - before.TotalAlloc; grew > 1<<20 {
		t.Errorf("slotsFromAmount(5e6) allocated %d bytes, want well under 1 MiB", grew)
	}
}
