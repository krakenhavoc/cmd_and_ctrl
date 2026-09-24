package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hideaway_view_test.go — the wire half of ADR 0091 (#1331). A card
// hidden by hideaway is a face-down exile whose one viewer is the
// hideaway permanent's controller (CR 702.75a): they read the card, the
// table reads a card back labelled with the public kind "hideaway", and
// nothing about the card's identity reaches anyone else.
func TestAHiddenCardIsReadableOnlyByTheSourcesController(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	src := game.NewCard("Wire Heights", me.ID)
	src.TypeLine = "Land"
	src.Controller = me.ID
	src.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	g.Battlefield.PushTop(src)
	hidden := game.NewCard("Wire Hidden Bolt", me.ID)
	hidden.TypeLine = "Instant"
	hidden.ManaCost = "{R}"
	me.Library.PushTop(hidden)
	g.WithWriteLock(func() {
		if _, err := g.ExileHiddenForEffect(game.ObjectRefOf(src), me.ID, hidden.InstanceID); err != nil {
			t.Fatalf("ExileHiddenForEffect: %v", err)
		}
	})

	find := func(v GameView) *CardView {
		for i := range v.Exile.Cards {
			if v.Exile.Cards[i].InstanceID == hidden.InstanceID.String() {
				return &v.Exile.Cards[i]
			}
		}
		return nil
	}
	mine := find(ViewOfGameFor(g, me.ID.String()))
	theirs := find(ViewOfGameFor(g, opp.ID.String()))
	if mine == nil || theirs == nil {
		t.Fatal("the hidden card is missing from a viewer's exile")
	}
	if mine.Name != "Wire Hidden Bolt" || !mine.FaceVisible {
		t.Errorf("controller reads %q, face_visible %v — want the card", mine.Name, mine.FaceVisible)
	}
	if theirs.Name != "" || theirs.FaceVisible {
		t.Errorf("opponent reads %q, face_visible %v — want a card back", theirs.Name, theirs.FaceVisible)
	}
	if theirs.FaceDownKind != string(game.FaceDownHidden) || !theirs.FaceDown {
		t.Errorf("opponent's view: face_down %v kind %q, want the public kind %q", theirs.FaceDown, theirs.FaceDownKind, game.FaceDownHidden)
	}
}
