package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #1530, #1968: a seat's trigger-order mode is visible to that seat and
// nobody else.
func TestTriggerOrderPreferenceIsPrivateToItsSeat(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if err := g.SetTriggerOrderPreference(a.ID, game.TriggerOrderAlways); err != nil {
		t.Fatal(err)
	}
	if err := g.SetTriggerOrderPreference(b.ID, game.TriggerOrderNever); err != nil {
		t.Fatal(err)
	}
	raw := ViewOfGame(g)
	for _, viewer := range []struct {
		id         string
		seeA, seeB string
	}{
		{a.ID.String(), "always", ""},
		{b.ID.String(), "", "never"},
		{SpectatorViewerID, "", ""},
	} {
		v := FilterViewFor(raw, viewer.id)
		if v.Seats[0].TriggerOrder != viewer.seeA || v.Seats[0].TriggerOrderAlwaysAsk != (viewer.seeA == "always") {
			t.Errorf("viewer %s sees seat A's mode as %q / %v", viewer.id, v.Seats[0].TriggerOrder, v.Seats[0].TriggerOrderAlwaysAsk)
		}
		if v.Seats[1].TriggerOrder != viewer.seeB || v.Seats[1].TriggerOrderAlwaysAsk {
			t.Errorf("viewer %s sees seat B's mode as %q / %v", viewer.id, v.Seats[1].TriggerOrder, v.Seats[1].TriggerOrderAlwaysAsk)
		}
	}
}
