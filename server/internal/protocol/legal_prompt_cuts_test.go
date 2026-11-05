package protocol

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// legal_prompt_cuts_test.go — #4765. A prompt the enumerator cut (a
// discard, a library search) reported its cut only in the full-list
// reply, and a frame whose only moves answer a prompt carries no
// legal_actions digest. So on a list under the wire's 48-move cap a seat
// reading the frame could not tell the list was cut. legal_prompt_cuts
// carries the prompt's cuts on the frame itself, own seat only.

func TestACutPromptUnderTheWireCapSaysSoOnTheFrame(t *testing.T) {
	g := busyTable(t, 0)
	active := g.Seats[g.Turn.ActiveSeat]
	active.Hand.Cards = nil
	for i := range 9 {
		put(active.Hand, active, beast(fmt.Sprintf("Beast %d", i), "{1}", 1, 1))
	}
	advanceTo(t, g, game.StepEnd)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	v := ViewOfGameFor(g, active.ID.String())
	// Nine cards, two to discard: C(9,2) = 36 answers, twelve listed.
	if v.LegalMovesTruncated || len(v.LegalMoves) != 12 {
		t.Fatalf("%d moves, truncated=%v; want the twelve discards under the wire cap", len(v.LegalMoves), v.LegalMovesTruncated)
	}
	if v.LegalActions != nil {
		t.Errorf("a frame of prompt answers grew a digest: %+v", v.LegalActions)
	}
	want := LegalCutView{Choice: legal.CleanupDiscardChoice, Cap: string(legal.CapPerSource), Omitted: 24}
	if len(v.LegalPromptCuts) != 1 || v.LegalPromptCuts[0] != want {
		t.Fatalf("legal_prompt_cuts = %+v; want [%+v]", v.LegalPromptCuts, want)
	}

	// Own seat only, like the list it describes.
	for _, p := range g.Seats {
		if p != active && ViewOfGameFor(g, p.ID.String()).LegalPromptCuts != nil {
			t.Errorf("seat %s got the active seat's prompt cuts", p.Name)
		}
	}
	if ViewOfGameFor(g, SpectatorViewerID).LegalPromptCuts != nil || ViewOfGameFor(g, "").LegalPromptCuts != nil {
		t.Error("a seatless viewer got the prompt cuts")
	}
}

// A card's own cut stays in its digest entry and is not repeated here.
func TestACutCardIsNotAPromptCut(t *testing.T) {
	g, spell, _ := removalTable(t, 7)
	active := g.Seats[g.Turn.ActiveSeat]
	v := ViewOfGameFor(g, active.ID.String())
	if v.LegalActions == nil || len(v.LegalActions.Sources[spell.String()].Truncated) == 0 {
		t.Fatal("the spell's cut is not in its digest entry")
	}
	if v.LegalPromptCuts != nil {
		t.Errorf("legal_prompt_cuts = %+v; a card's cut is not a prompt's", v.LegalPromptCuts)
	}
}
