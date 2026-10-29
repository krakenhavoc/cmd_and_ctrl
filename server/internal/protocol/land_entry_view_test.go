package protocol

import "testing"

func TestLandEntryFactsComeFromCatalogShapes(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle, entry string
		bounce              bool
	}{
		{"Chamber", "046f5783-cc7b-416a-8cf6-2bcef9c2cc1a", "tapped", true},
		{"Path", "b473e293-59e3-4e04-acf2-622604aeb25f", "tapped", false},
		{"Tower", "0895c9b7-ae7d-4bb3-af17-3b75deb50a25", "untapped", false},
		{"Conditional checkland", "027dd013-baa7-4111-b3c9-f4d1414e9c45", "", false},
	} {
		c := rowsFixtureCard(tc.name, "Land", tc.oracle, me.ID)
		c.AddKnower(me.ID)
		g.WithWriteLock(func() { me.Hand.PushTop(c) })
		g.BumpLayerVersionForTest()
		v := findCardView(t, seatZones(ViewOfGameFor(g, me.ID.String()), me.ID.String()), c.InstanceID)
		entry, bounce := "", false
		if v.Purpose != nil {
			entry, bounce = v.Purpose.LandEntry, v.Purpose.ReturnsLandOnEntry
		}
		if entry != tc.entry || bounce != tc.bounce {
			t.Errorf("%s: entry %q bounce %v, want %q %v", tc.name, entry, bounce, tc.entry, tc.bounce)
		}
	}
}
