package model

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0114 §7: the model tiers read the Ring off the board text — one
// line per emblem with the Ring's count, and a "Ring-bearer" tag on the
// card.
func TestBoardTextShowsTheRingAndTheRingBearer(t *testing.T) {
	ring := protocol.EmblemView{
		Label: "The Ring",
		Text:  "Your Ring-bearer is legendary and can't be blocked by creatures with greater power.\nWhenever your Ring-bearer attacks, draw a card, then discard a card.",
		Level: 3,
	}
	got := emblemLine(ring)
	want := "The Ring (tempted 3 times): Your Ring-bearer is legendary and can't be blocked by creatures with greater power. Whenever your Ring-bearer attacks, draw a card, then discard a card."
	if got != want {
		t.Errorf("emblem line\n got %q\nwant %q", got, want)
	}
	ring.Level = 1
	if got := emblemLine(ring); !strings.HasPrefix(got, "The Ring (tempted 1 time): ") {
		t.Errorf("one temptation reads %q", got)
	}
	if got := emblemLine(protocol.EmblemView{Label: "Elspeth, Sun's Champion emblem", Text: "Creatures you control get +2/+2 and have flying."}); got != "Elspeth, Sun's Champion emblem: Creatures you control get +2/+2 and have flying." {
		t.Errorf("an ordinary emblem reads %q", got)
	}

	c := &protocol.CardView{Name: "Nazgûl", TypeLine: "Legendary Creature — Zombie Wraith Knight", Power: 1, Toughness: 2, RingBearer: true}
	if got := permanentLabel(c); !strings.Contains(got, "Ring-bearer") {
		t.Errorf("the Ring-bearer's label %q does not say so", got)
	}
}
