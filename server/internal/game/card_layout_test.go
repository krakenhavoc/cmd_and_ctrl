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
		"Move it into the bool block at the end of Card and leave a pointer comment where it was.",
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

// paddingGap is one run of alignment padding between two fields: size
// bytes between the end of the field named after and the start of the
// field described by before.
type paddingGap struct {
	after  string
	before string
	size   uintptr
}

// structPadding walks a struct type's fields in declaration order and
// returns every gap between one field's end and the next field's
// offset (interior), and separately the padding after the last field
// (tail).
func structPadding(typ reflect.Type) (interior []paddingGap, tail uintptr) {
	var end uintptr
	prev := "start of struct"
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if gap := f.Offset - end; gap > 0 {
			interior = append(interior, paddingGap{
				after:  prev,
				before: fmt.Sprintf("%s (%s, align %d)", f.Name, f.Type, f.Type.Align()),
				size:   gap,
			})
		}
		end = f.Offset + f.Type.Size()
		prev = f.Name
	}
	return interior, typ.Size() - end
}
