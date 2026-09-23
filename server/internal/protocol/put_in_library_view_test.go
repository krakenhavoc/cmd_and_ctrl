package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestPutInLibraryProjectsPlacementAndOnlyToTheChooser — ADR 0088's
// ordered placement on the wire: the chooser gets the cards (in the
// order the prompt lists them) and the placement that tells the dialog
// which lanes to open; another seat gets the prompt, the count and the
// placement, and not one handle on the cards — Brainstorm's put-back is
// two cards out of a HAND.
func TestPutInLibraryProjectsPlacementAndOnlyToTheChooser(t *testing.T) {
	g := buildThreeSeatGame(t)
	me, other := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	var names []string
	g.WithWriteLock(func() {
		for _, name := range []string{"Putback Alpha", "Putback Beta"} {
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
			ids = append(ids, id)
			names = append(names, name)
		}
		if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   me.ID,
			Cards:     ids,
			From:      game.ZoneHand,
			Placement: game.LibraryPlaceTop,
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})

	kind := string(game.PendingChoicePutInLibrary)
	mine := choiceFor(t, ViewOfGameFor(g, me.ID.String()), kind)
	if mine.Placement != "top" {
		t.Errorf("placement %q on the wire, want \"top\"", mine.Placement)
	}
	if len(mine.Options) != 2 || mine.Options[0].Name != names[0] || mine.Options[1].Name != names[1] {
		t.Fatalf("the chooser's options %+v, want the two cards in order", mine.Options)
	}

	theirs := ViewOfGameFor(g, other.ID.String())
	tc := choiceFor(t, theirs, kind)
	if len(tc.Options) != 0 {
		t.Errorf("another seat got %d options out of somebody's hand", len(tc.Options))
	}
	if tc.Count != 2 || tc.Placement != "top" {
		t.Errorf("count %d / placement %q stay public", tc.Count, tc.Placement)
	}
	buf, err := json.Marshal(theirs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for i, id := range ids {
		if strings.Contains(string(buf), id.String()) || strings.Contains(string(buf), names[i]) {
			t.Errorf("card %d reached another seat's frame", i)
		}
	}
}
