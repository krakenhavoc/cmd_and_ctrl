package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0191 §8: the seat's own sideboard is on its board text so an agent
// at the table can read what a wish could bring in; an opponent's is a
// count.
func TestRenderShowsTheOwnSideboardAndAnOpponentsCount(t *testing.T) {
	v := view()
	v.Seats[0].Sideboard = []protocol.CardView{
		{Name: "Burning Wish", ManaCost: "{1}{R}", TypeLine: "Sorcery"},
		{Name: "Lightning Bolt", ManaCost: "{R}", TypeLine: "Instant"},
	}
	v.Seats[0].SideboardCount = 2
	v.Seats[1].SideboardCount = 5
	got := boardtext.Render(v, "a", boardtext.Options{})
	if !strings.Contains(got, "  your sideboard (outside the game): Burning Wish {1}{R}") || !strings.Contains(got, "Lightning Bolt") {
		t.Errorf("the seat's own sideboard is not described:\n%s", got)
	}
	if !strings.Contains(got, "  sideboard (outside the game): 5 cards") {
		t.Errorf("an opponent's sideboard is not a count:\n%s", got)
	}
	if strings.Count(got, "sideboard (outside the game)") != 2 {
		t.Errorf("want exactly the own line and the opponent's count:\n%s", got)
	}
	// No sideboard, no line.
	if bare := boardtext.Render(view(), "a", boardtext.Options{}); strings.Contains(bare, "sideboard") {
		t.Errorf("a seat with none shows a line:\n%s", bare)
	}
}
