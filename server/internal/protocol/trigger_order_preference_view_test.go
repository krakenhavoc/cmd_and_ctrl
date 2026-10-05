package protocol

import "testing"

// #1530: a seat's "always ask me to order my triggers" preference is
// visible to that seat and nobody else.
func TestTriggerOrderPreferenceIsPrivateToItsSeat(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if err := g.SetTriggerOrderPreference(a.ID, true); err != nil {
		t.Fatal(err)
	}
	raw := ViewOfGame(g)
	for _, viewer := range []struct {
		id         string
		seeA, seeB bool
	}{
		{a.ID.String(), true, false},
		{b.ID.String(), false, false},
		{SpectatorViewerID, false, false},
	} {
		v := FilterViewFor(raw, viewer.id)
		if v.Seats[0].TriggerOrderAlwaysAsk != viewer.seeA {
			t.Errorf("viewer %s sees seat A's preference as %v", viewer.id, v.Seats[0].TriggerOrderAlwaysAsk)
		}
		if v.Seats[1].TriggerOrderAlwaysAsk != viewer.seeB {
			t.Errorf("viewer %s sees seat B's preference as %v", viewer.id, v.Seats[1].TriggerOrderAlwaysAsk)
		}
	}
}
