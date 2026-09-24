package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestPutInLibraryLookAtAnotherPlayersLibraryOnlyTheLookerSees — #1298.
// Jace, the Mind Sculptor's +2 and Portent look at ANOTHER player's
// library: the looker's wire carries the cards; the library's OWNER's
// wire carries the prompt, its count and its shape, and not one handle
// on the cards — CR 401.2 keeps a library face down to its owner too.
// The top lane's exact count and depth are public, like the placement.
func TestPutInLibraryLookAtAnotherPlayersLibraryOnlyTheLookerSees(t *testing.T) {
	g := buildThreeSeatGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
	names := []string{"Peeked Alpha", "Peeked Beta", "Peeked Gamma"}
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		for i := len(names) - 1; i >= 0; i-- {
			id := uuid.New()
			owner.Library.PushTop(game.Card{InstanceID: id, Name: names[i], TypeLine: "Instant", Owner: owner.ID, Controller: owner.ID})
			ids = append([]uuid.UUID{id}, ids...)
		}
		looked := g.LookAtTopOfPlayersLibraryForEffect(me.ID, owner.ID, 3)
		if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
			Chooser:   me.ID,
			Cards:     looked,
			From:      game.ZoneLibrary,
			Placement: game.LibraryPlaceTopOrBottom,
			TopCount:  1,
			TopDepth:  2,
		}); err != nil {
			t.Fatalf("queue: %v", err)
		}
	})

	kind := string(game.PendingChoicePutInLibrary)
	mine := choiceFor(t, ViewOfGameFor(g, me.ID.String()), kind)
	if len(mine.Options) != 3 || mine.Options[0].Name != names[0] {
		t.Fatalf("the looker's options %+v, want the owner's top three, top-first", mine.Options)
	}
	if mine.TopCount != 1 || mine.TopDepth != 2 || mine.Placement != "top_or_bottom" {
		t.Errorf("top_count %d / top_depth %d / placement %q on the wire", mine.TopCount, mine.TopDepth, mine.Placement)
	}

	for _, seat := range []*game.Player{owner, other} {
		view := ViewOfGameFor(g, seat.ID.String())
		theirs := choiceFor(t, view, kind)
		if len(theirs.Options) != 0 {
			t.Errorf("seat %s got %d options out of a library it may not see", seat.Name, len(theirs.Options))
		}
		if theirs.Count != 3 || theirs.TopCount != 1 || theirs.TopDepth != 2 {
			t.Errorf("the prompt's shape stays public: count %d, top_count %d, top_depth %d",
				theirs.Count, theirs.TopCount, theirs.TopDepth)
		}
		buf, err := json.Marshal(view)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		for i, id := range ids {
			if strings.Contains(string(buf), names[i]) {
				t.Errorf("card %d's face reached seat %s's frame", i, seat.Name)
			}
			// The owner's own library already carries its cards' IDs
			// (face down) on the owner's wire; a third seat's must not
			// gain a handle on them through the prompt.
			if seat == other && strings.Contains(string(buf), id.String()) {
				t.Errorf("card %d's ID reached a third seat's frame", i)
			}
		}
	}
}
