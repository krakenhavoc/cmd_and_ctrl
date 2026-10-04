package protocol

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// land_drop_frame_test.go — #2203. The hand's ready ring on a land is
// the frame's `legal_actions` digest, and the client's fallback for a
// frame with no list reads the seat's `land_drops_per_turn` and
// `lands_played_this_turn`. Both are pinned on one frame: once the
// turn's land drop is spent, no hand land is in the digest and the seat
// says the allowance is used.
func TestFrameAfterTheLandDropOffersNoHandLand(t *testing.T) {
	g := busyTable(t, 1)
	active := g.Seats[g.Turn.ActiveSeat]

	handLandKinds := func(v GameView) []string {
		var out []string
		if v.LegalActions == nil {
			return nil
		}
		for _, c := range active.Hand.Cards {
			if !c.IsLand() {
				continue
			}
			if s := v.LegalActions.Sources[c.InstanceID.String()]; s != nil && slices.Contains(s.Kinds, legal.KindLand) {
				out = append(out, c.Name)
			}
		}
		return out
	}
	seatOf := func(v GameView) *PlayerView {
		for i := range v.Seats {
			if v.Seats[i].ID == active.ID.String() {
				return &v.Seats[i]
			}
		}
		t.Fatal("active seat missing from its own frame")
		return nil
	}

	before := ViewOfGameFor(g, active.ID.String())
	if got := handLandKinds(before); len(got) != 3 {
		t.Fatalf("before the land drop: want 3 playable hand lands in the digest, got %v", got)
	}
	if s := seatOf(before); s.LandDropsPerTurn != 1 || s.LandsPlayedThisTurn != 0 {
		t.Fatalf("before the land drop: seat says %d of %d played, want 0 of 1", s.LandsPlayedThisTurn, s.LandDropsPerTurn)
	}

	spendLandDrop(t, g, active)

	after := ViewOfGameFor(g, active.ID.String())
	if after.LegalActions == nil {
		t.Fatal("after the land drop the seat still holds priority, so the frame should carry a digest")
	}
	if got := handLandKinds(after); len(got) != 0 {
		t.Errorf("after the land drop: hand lands still in the digest: %v", got)
	}
	if s := seatOf(after); s.LandDropsPerTurn != 1 || s.LandsPlayedThisTurn != 1 {
		t.Errorf("after the land drop: seat says %d of %d played, want 1 of 1", s.LandsPlayedThisTurn, s.LandDropsPerTurn)
	}
}
