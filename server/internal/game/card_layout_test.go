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

// maxCardPadding is the most alignment padding Card may carry. The
// bool block leaves a few bytes of tail padding (it is rarely a
// multiple of 8); one stranded bool costs 7 more and trips it.
const maxCardPadding = 8

func TestCardAlignmentPaddingStaysSmall(t *testing.T) {
	gaps, total := structPadding(reflect.TypeOf(Card{}))
	if total <= maxCardPadding {
		return
	}
	var b strings.Builder
	for _, g := range gaps {
		fmt.Fprintf(&b, "\n  %d B after %s, before %s", g.size, g.after, g.before)
	}
	t.Errorf("game.Card carries %d B of alignment padding, want <= %d B:%s\n"+
		"A bool (or other small field) is probably sitting alone before a wider field; "+
		"uuid.UUID aligns to 1, so it may be a few fields above the gap. "+
		"Move it into the bool block at the end of Card and leave a pointer comment where it was.",
		total, maxCardPadding, b.String())
}

// TestStructPaddingSeesAStrandedBool checks the walker itself, so the
// guard above cannot pass by measuring nothing.
func TestStructPaddingSeesAStrandedBool(t *testing.T) {
	type stranded struct {
		A bool
		B int64
		C bool
	}
	gaps, total := structPadding(reflect.TypeOf(stranded{}))
	want := []paddingGap{
		{after: "A", before: "B (int64, align 8)", size: 7},
		{after: "C", before: "end of struct", size: 7},
	}
	if total != 14 || !reflect.DeepEqual(gaps, want) {
		t.Fatalf("structPadding(stranded) = %+v, %d B; want %+v, 14 B", gaps, total, want)
	}
}

// paddingGap is one run of alignment padding inside a struct: size
// bytes between the end of the field named after and the start of the
// field described by before ("end of struct" for tail padding).
type paddingGap struct {
	after  string
	before string
	size   uintptr
}

// structPadding walks a struct type's fields in declaration order and
// returns every gap between one field's end and the next field's
// offset, plus the tail padding after the last field.
func structPadding(typ reflect.Type) (gaps []paddingGap, total uintptr) {
	var end uintptr
	prev := "start of struct"
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if gap := f.Offset - end; gap > 0 {
			gaps = append(gaps, paddingGap{
				after:  prev,
				before: fmt.Sprintf("%s (%s, align %d)", f.Name, f.Type, f.Type.Align()),
				size:   gap,
			})
			total += gap
		}
		end = f.Offset + f.Type.Size()
		prev = f.Name
	}
	if gap := typ.Size() - end; gap > 0 {
		gaps = append(gaps, paddingGap{after: prev, before: "end of struct", size: gap})
		total += gap
	}
	return gaps, total
}
