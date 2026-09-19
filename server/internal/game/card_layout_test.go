package game

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// card_layout_test.go guards game.Card's memory layout.
//
// Card is the most-copied struct in the engine (every zone move,
// every clone, every snapshot), and its fields are grouped by the
// sprint and mechanic that added them rather than by alignment. That
// grouping is worth keeping, but it has one cost: a lone bool between
// two 8-byte fields strands 7 bytes of padding. Six of them did,
// which cost Card 32 bytes (#35).
//
// The fix was to move every bool into one block at the end of the
// struct. This test is what keeps it that way, in place of
// govet/fieldalignment, which stays disabled repo-wide
// (server/.golangci.yml). It deliberately does NOT pin
// unsafe.Sizeof(Card{}): adding a field must not fail it, and
// stranding a bool must.
//
// The rule is zero padding BETWEEN fields; only tail padding after
// the last field is allowed, since the bool block is rarely a
// multiple of 8. An earlier "total padding <= 8 B" rule missed a
// stranded bool whenever the block's own tail padding shrank to make
// room for it (TestStructPaddingSeesWhatATotalMissed).
//
// The walker RECURSES into struct-valued (non-pointer) fields (#633):
// a bool stranded inside an embedded struct — Card.AttachedTo
// (TargetRef), Card.Provenance (CastProvenance), or a future one —
// used to be invisible, because the field it strands padding inside
// of is measured as one opaque block from the top level. A gap found
// this way is reported with its full dotted path, e.g.
// "Provenance.FromZone", so the failure names the struct to fix.
// Array-kind fields (uuid.UUID is [16]byte) are NOT recursed into —
// there is no second field to strand anything between — and neither
// are pointer fields, since what they point at is not part of this
// struct's own layout.

func TestCardHasNoInteriorPadding(t *testing.T) {
	interior, _ := structPadding(reflect.TypeOf(Card{}))
	if len(interior) == 0 {
		return
	}
	var b strings.Builder
	for _, g := range interior {
		fmt.Fprintf(&b, "\n  %d B after %s, before %s", g.size, g.after, g.before)
	}
	t.Errorf("game.Card has alignment padding between fields; want none:%s\n"+
		"A bool (or other small field) is probably sitting alone before a wider field; "+
		"uuid.UUID aligns to 1, so it may be a few fields above the gap. "+
		"Move it into the bool block at the end of Card (or the bool block of "+
		"whichever nested struct the path names) and leave a pointer comment "+
		"where it was.",
		b.String())
}

// TestStructPaddingSeesAStrandedBool checks the walker itself, so the
// guard above cannot pass by measuring nothing.
func TestStructPaddingSeesAStrandedBool(t *testing.T) {
	type stranded struct {
		A bool
		B int64
		C bool
	}
	interior, tail := structPadding(reflect.TypeOf(stranded{}))
	want := []paddingGap{{after: "A", before: "B (int64, align 8)", size: 7}}
	if !reflect.DeepEqual(interior, want) || tail != 7 {
		t.Fatalf("structPadding(stranded) = %+v, tail %d B; want %+v, tail 7 B", interior, tail, want)
	}
}

// TestStructPaddingSeesWhatATotalMissed is the case the old "total
// padding <= 8 B" rule let through: Card's six-bool block grown by two
// more bools (a multiple of 8, so no tail padding), plus one bool
// stranded mid-struct. Total padding is 7 B, under the old limit;
// the interior rule still names the gap.
func TestStructPaddingSeesWhatATotalMissed(t *testing.T) {
	type grownBlockOneStranded struct {
		Power    int
		Stranded bool
		Counters map[string]int

		NeedsEffect              bool
		Tapped                   bool
		IsCommander              bool
		FaceDown                 bool
		SummonedThisTurn         bool
		MarkedLethalByDeathtouch bool
		ExtraA                   bool
		ExtraB                   bool
	}
	interior, tail := structPadding(reflect.TypeOf(grownBlockOneStranded{}))
	var total uintptr
	for _, g := range interior {
		total += g.size
	}
	total += tail
	if total > 8 {
		t.Fatalf("total padding = %d B; this fixture only means something if the old <= 8 B rule would pass it", total)
	}
	want := []paddingGap{{after: "Stranded", before: "Counters (map[string]int, align 8)", size: 7}}
	if !reflect.DeepEqual(interior, want) || tail != 0 {
		t.Fatalf("structPadding = %+v, tail %d B; want %+v, tail 0 B", interior, tail, want)
	}
}

// TestStructPaddingRecursesIntoNestedStructs is the walker's own
// self-check for #633: a stranded bool one level down, inside a
// struct-valued (non-pointer) field, must be reported by its full
// dotted path rather than being absorbed into the opaque block the
// outer field used to be measured as.
func TestStructPaddingRecursesIntoNestedStructs(t *testing.T) {
	type nested struct {
		A bool
		B int64
	}
	type outer struct {
		X      int64
		Nested nested
		Y      bool
	}
	interior, _ := structPadding(reflect.TypeOf(outer{}))
	want := []paddingGap{{after: "Nested.A", before: "Nested.B (int64, align 8)", size: 7}}
	if !reflect.DeepEqual(interior, want) {
		t.Fatalf("structPadding(outer) = %+v; want %+v (the nested gap, reported by dotted path)", interior, want)
	}
}

// TestStructPaddingDoesNotRecurseIntoPointersOrArrays checks the two
// deliberate non-recursion cases: a pointer field's target is not
// part of THIS struct's layout, and an array-kind field (uuid.UUID is
// [16]byte) has no second field to strand anything between, so
// recursing into either would only ever report nothing.
func TestStructPaddingDoesNotRecurseIntoPointersOrArrays(t *testing.T) {
	type strandedInsidePointee struct {
		A bool
		B int64
	}
	type outer struct {
		Arr  [16]byte
		Ptr  *strandedInsidePointee
		Tail bool
	}
	interior, _ := structPadding(reflect.TypeOf(outer{}))
	if len(interior) != 0 {
		t.Fatalf("structPadding(outer) = %+v; want no interior padding — Arr and Ptr are both leaves", interior)
	}
}

// paddingGap is one run of alignment padding between two fields: size
// bytes between the end of the field named after and the start of the
// field described by before. after and before are dotted paths
// ("Provenance.FromZone") when the gap was found inside a nested,
// struct-valued field rather than at the top level.
type paddingGap struct {
	after  string
	before string
	size   uintptr
}

// structPadding walks a struct type's fields in declaration order and
// returns every gap between one field's end and the next field's
// offset (interior), and separately the padding after the last field
// (tail). It recurses into struct-valued (non-pointer) fields, so a
// bool stranded inside an embedded struct is reported too, by its
// full dotted path — see the file header (#633). Pointer fields and
// array-kind fields (uuid.UUID) are leaves: what a pointer targets is
// not this struct's own layout, and an array has no second field to
// strand anything between.
//
// Interior gaps found inside a nested field are that nested struct's
// OWN layout question — they do not depend on where the field sits in
// the outer struct — so the recursive call needs no offset math, only
// a path prefix for the field names it reports.
func structPadding(typ reflect.Type) (interior []paddingGap, tail uintptr) {
	return structPaddingPath(typ, "")
}

func structPaddingPath(typ reflect.Type, pathPrefix string) (interior []paddingGap, tail uintptr) {
	var end uintptr
	prev := "start of struct"
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		name := f.Name
		if pathPrefix != "" {
			name = pathPrefix + "." + f.Name
		}
		if gap := f.Offset - end; gap > 0 {
			interior = append(interior, paddingGap{
				after:  prev,
				before: fmt.Sprintf("%s (%s, align %d)", name, f.Type, f.Type.Align()),
				size:   gap,
			})
		}
		end = f.Offset + f.Type.Size()
		prev = name
		if f.Type.Kind() == reflect.Struct {
			nestedInterior, _ := structPaddingPath(f.Type, name)
			interior = append(interior, nestedInterior...)
		}
	}
	return interior, typ.Size() - end
}
