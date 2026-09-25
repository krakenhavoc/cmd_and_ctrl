package game

import "testing"

// replacement_id_test.go covers #801: catalog ReplacementEffectIDs
// pack a battlefield index and a slot into one number, and that
// packing used to be a literal in the gather pass plus a second,
// hand-rolled literal in ReplacementOptionMetaForEffect. One
// encode/decode pair now owns it, and this is the round trip.

func TestCatalogReplacementIDRoundTrip(t *testing.T) {
	cases := []struct {
		name    string
		cardIdx int
		slot    int
		wantOK  bool
	}{
		{"first card, first slot", 0, 0, true},
		{"first card, last slot", 0, MaxCatalogReplacementSlots - 1, true},
		{"second card, first slot", 1, 0, true},
		{"a realistic battlefield", 37, 2, true},
		{"a very large battlefield", 1 << 20, 255, true},
		{"slot past the budget", 0, MaxCatalogReplacementSlots, false},
		{"slot far past the budget", 3, 900, false},
		{"negative slot", 0, -1, false},
		{"negative card index", -1, 0, false},
		// A card index high enough to mint into the
		// self-replacement range would name the wrong registry.
		{"card index past the catalog range", int(selfReplacementIDBase / MaxCatalogReplacementSlots), 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, ok := encodeCatalogReplacementID(tc.cardIdx, tc.slot)
			if ok != tc.wantOK {
				t.Fatalf("encodeCatalogReplacementID(%d, %d) ok = %v, want %v", tc.cardIdx, tc.slot, ok, tc.wantOK)
			}
			if !ok {
				if id != 0 {
					t.Errorf("refused encode returned id %d, want 0", id)
				}
				return
			}
			if id == 0 {
				t.Fatalf("encodeCatalogReplacementID(%d, %d) = 0, which means \"no ID\"", tc.cardIdx, tc.slot)
			}
			gotCard, gotSlot, gotOK := decodeCatalogReplacementID(id)
			if !gotOK {
				t.Fatalf("decodeCatalogReplacementID(%d) ok = false, want true", id)
			}
			if gotCard != tc.cardIdx || gotSlot != tc.slot {
				t.Errorf("round trip = (%d, %d), want (%d, %d)", gotCard, gotSlot, tc.cardIdx, tc.slot)
			}
		})
	}
}

// TestDecodeCatalogReplacementIDRejectsOtherRanges is the defensive
// half: an ID minted by one of the registries ABOVE the catalog must
// not decode to a battlefield index, or a stale prompt would name
// whatever card happens to sit there.
func TestDecodeCatalogReplacementIDRejectsOtherRanges(t *testing.T) {
	cases := []struct {
		name string
		id   ReplacementEffectID
	}{
		{"zero is not an ID", 0},
		{"self-replacement", selfReplacementIDBase},
		{"self-replacement, slot 3", selfReplacementIDBase + 3},
		{"scoped replacement", scopedReplacementIDBase},
		{"test-injected", testReplacementIDBase},
		{"built-in", builtinReplacementIDBase},
		{"built-in, index 2", builtinReplacementIDBase + 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, ok := decodeCatalogReplacementID(tc.id); ok {
				t.Errorf("decodeCatalogReplacementID(%d) ok = true, want false", tc.id)
			}
		})
	}
}

// TestCatalogReplacementIDsAreDistinctPerSlotAndCard pins the
// property the packing exists for: no two (card, slot) pairs share an
// ID, which is what would let one permanent's prompt entry reorder
// another's effect.
func TestCatalogReplacementIDsAreDistinctPerSlotAndCard(t *testing.T) {
	seen := make(map[ReplacementEffectID]string)
	for cardIdx := 0; cardIdx < 40; cardIdx++ {
		for slot := 0; slot < 4; slot++ {
			id, ok := encodeCatalogReplacementID(cardIdx, slot)
			if !ok {
				t.Fatalf("encodeCatalogReplacementID(%d, %d) refused a representable pair", cardIdx, slot)
			}
			where := "card " + ReplacementEffectIDToString(ReplacementEffectID(cardIdx)) +
				" slot " + ReplacementEffectIDToString(ReplacementEffectID(slot))
			if prev, dup := seen[id]; dup {
				t.Fatalf("id %d minted for both %s and %s", id, prev, where)
			}
			seen[id] = where
		}
	}
}
