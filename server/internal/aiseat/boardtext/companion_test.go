package boardtext_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0172 §7: every seat's revealed companion is on the board text,
// the seat's own with its cost and type.
func TestRenderShowsEveryRevealedCompanion(t *testing.T) {
	v := view()
	v.Seats[0].Companion = &protocol.CardView{Name: "Jegantha, the Wellspring", ManaCost: "{4}{R/G}", TypeLine: "Legendary Creature — Elemental Elk", IsCompanion: true}
	v.Seats[1].Companion = &protocol.CardView{Name: "Jegantha, the Wellspring", ManaCost: "{4}{R/G}", TypeLine: "Legendary Creature — Elemental Elk", IsCompanion: true}
	got := boardtext.Render(v, "a", boardtext.Options{})
	if n := strings.Count(got, "  companion (outside the game): Jegantha, the Wellspring"); n != 2 {
		t.Errorf("%d companion lines, want one per seat:\n%s", n, got)
	}
	if !strings.Contains(got, "  companion (outside the game): Jegantha, the Wellspring {4}{R/G}") {
		t.Errorf("the seat's own companion is not described:\n%s", got)
	}
}
