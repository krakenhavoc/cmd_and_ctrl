package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestNoUntapViewStampsMarkersAndFaceDownPrivacy(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	staticID, markerID, hiddenID := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: staticID, Name: "Static target", Owner: owner.ID, Controller: owner.ID,
			Tapped: true,
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: markerID, Name: "Marked target", Owner: owner.ID, Controller: owner.ID,
			NextUntapSkips: []game.UntapSkip{{}, {Player: other.ID}, {Player: other.ID}, {Player: uuid.New()}},
		})
		g.Battlefield.PushTop(game.Card{
			InstanceID: hiddenID, Name: "Hidden target", Owner: owner.ID, Controller: owner.ID,
			FaceDown: true, NextUntapSkips: []game.UntapSkip{{Player: other.ID}},
		})
	})

	prev := game.CatalogUntapStepRestrictions
	game.CatalogUntapStepRestrictions = func(key string) []game.UntapStepRestriction {
		if key != "restriction-source" {
			return nil
		}
		return []game.UntapStepRestriction{{Restricts: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.InstanceID == staticID || target.InstanceID == hiddenID
		}}}
	}
	t.Cleanup(func() { game.CatalogUntapStepRestrictions = prev })
	// The restriction source is public battlefield state and is separate from
	// the two cards being projected, so the test exercises the game predicate.
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Restriction source", OracleID: "restriction-source",
			Owner: owner.ID, Controller: owner.ID,
		})
	})

	find := func(v GameView, id uuid.UUID) CardView {
		t.Helper()
		for _, card := range v.Battlefield.Cards {
			if card.InstanceID == id.String() {
				return card
			}
		}
		t.Fatalf("card %s missing from battlefield", id)
		return CardView{}
	}
	v := ViewOfGameFor(g, owner.ID.String())
	static := find(v, staticID)
	if static.NoUntap == nil || !static.NoUntap.Static || len(static.NoUntap.Next) != 0 {
		t.Fatalf("static no_untap = %+v, want static only", static.NoUntap)
	}
	marker := find(v, markerID)
	if marker.NoUntap == nil || marker.NoUntap.Static || len(marker.NoUntap.Next) != 2 ||
		marker.NoUntap.Next[0] != owner.ID.String() || marker.NoUntap.Next[1] != other.ID.String() {
		t.Fatalf("marker no_untap = %+v, want deduplicated controller and other IDs", marker.NoUntap)
	}
	hidden := find(v, hiddenID)
	if hidden.NoUntap == nil || hidden.NoUntap.Static || len(hidden.NoUntap.Next) != 1 || hidden.NoUntap.Next[0] != other.ID.String() {
		t.Fatalf("face-down no_untap = %+v, want next marker and no static bit", hidden.NoUntap)
	}
}

// #1313: a live "for as long as" hold projects as static — also on a
// face-down permanent, because the hold comes from another object —
// and never as a next-step entry. An expired hold projects nothing.
func TestNoUntapViewProjectsHoldsAsStatic(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	sourceID, heldID, hiddenID, staleID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: sourceID, Name: "Source", Owner: owner.ID, Controller: owner.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: heldID, Name: "Held", Owner: other.ID, Controller: other.ID, Tapped: true})
		g.Battlefield.PushTop(game.Card{InstanceID: hiddenID, Name: "Hidden", Owner: other.ID, Controller: other.ID, Tapped: true, FaceDown: true})
		g.Battlefield.PushTop(game.Card{InstanceID: staleID, Name: "Stale", Owner: other.ID, Controller: other.ID, Tapped: true})
		live, ok := g.ForAsLongAsYouControlDuration(sourceID, owner.ID)
		if !ok {
			t.Fatal("duration did not start")
		}
		stale := live
		stale.Player = other.ID // "for as long as other controls Source": already false
		for _, id := range []uuid.UUID{heldID, hiddenID} {
			_ = g.HoldUntappedForEffect(id, live)
		}
		_ = g.HoldUntappedForEffect(staleID, stale)
	})

	v := ViewOfGameFor(g, other.ID.String())
	byID := map[string]CardView{}
	for _, c := range v.Battlefield.Cards {
		byID[c.InstanceID] = c
	}
	for _, id := range []uuid.UUID{heldID, hiddenID} {
		got := byID[id.String()].NoUntap
		if got == nil || !got.Static || len(got.Next) != 0 {
			t.Errorf("%s no_untap = %+v, want static only", id, got)
		}
	}
	if got := byID[staleID.String()].NoUntap; got != nil {
		t.Errorf("expired hold projected no_untap = %+v, want none", got)
	}
}
