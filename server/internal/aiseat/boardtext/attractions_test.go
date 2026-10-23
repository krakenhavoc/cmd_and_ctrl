package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0207: an agent at the table reads every seat's Attraction deck as a
// count (it is face down, its owner's too), every seat's junkyard by name
// (it is public), and an Attraction's lit numbers on the battlefield.
func TestRenderDescribesAttractionPiles(t *testing.T) {
	v := view()
	v.Seats[0].AttractionDeckCount = 8
	v.Seats[1].AttractionDeckCount = 10
	v.Seats[1].Junkyard = []protocol.CardView{{Name: "Information Booth", TypeLine: "Artifact — Attraction"}}
	v.Battlefield.Cards = append(v.Battlefield.Cards, protocol.CardView{
		InstanceID: "attr", Name: "Spinny Ride", TypeLine: "Artifact — Attraction",
		Owner: "a", Controller: "a", AttractionLights: []int{2, 3, 6},
	})
	got := boardtext.Render(v, "a", boardtext.Options{})
	for _, want := range []string{
		"  attraction deck (face down, command zone): 8 cards",
		"  attraction deck (face down, command zone): 10 cards",
		"  junkyard (command zone): Information Booth",
		"Spinny Ride (lit 2, 3, 6)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "junkyard") != 1 {
		t.Errorf("want one junkyard line, for the seat that has one:\n%s", got)
	}
	if bare := boardtext.Render(view(), "a", boardtext.Options{}); strings.Contains(bare, "attraction deck") || strings.Contains(bare, "junkyard") {
		t.Errorf("a table with no Attractions shows a line:\n%s", bare)
	}
}
