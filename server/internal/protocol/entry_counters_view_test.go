package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEntryCountersComeOnlyFromReliableCatalogShapes(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle, kind string
		amount             int
		x                  bool
	}{
		{"Hydra", "7bd36106-04fe-481f-b16e-e076dcbb183b", "+1/+1", 4, false},
		{"Mikaeus", "82f3faa8-39fa-450b-843f-d60a4c36d8f7", "+1/+1", 0, true},
		{"Krasis", "6bd872b2-5c40-4e11-9a7f-0136a51b0642", "+1/+1", 0, true},
		{"Gemstone Mine", "0c828f10-4775-492f-9224-1e2814ad2cad", "mining", 3, false},
		{"Gnarlbark Elm", "a848ea46-347d-49c7-85c8-023cf836d0b3", "-1/-1", 2, false},
		{"Board-dependent Colossus", "fd62ad01-601f-4250-bc2a-8ef3982e45c4", "", 0, false},
		{"Conditional kicker X", "fbeafc81-1ddb-4b54-8f86-e6afd5f81db8", "", 0, false},
		{"Multikick", "0a79237e-0811-4a8a-bd4d-db3ca91bff22", "", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := rowsFixtureCard("Unrelated fixture name", "Creature", tc.oracle, me.ID)
			c.AddKnower(me.ID)
			g.WithWriteLock(func() { me.Hand.PushTop(c) })
			g.BumpLayerVersionForTest()
			v := findCardView(t, seatZones(ViewOfGameFor(g, me.ID.String()), me.ID.String()), c.InstanceID)
			var entries *[]EntryCounterView
			if v.Purpose != nil {
				entries = v.Purpose.EntryCounters
			}
			if tc.kind == "" {
				if entries != nil {
					t.Fatalf("conditional count projected: %+v", *entries)
				}
				return
			}
			want := EntryCounterView{Kind: tc.kind, Amount: tc.amount, AmountIsX: tc.x}
			if entries == nil || len(*entries) != 1 || (*entries)[0] != want {
				t.Fatalf("entry counters %+v, want %+v", entries, want)
			}
			encoded, err := json.Marshal(v.Purpose)
			if err != nil || !strings.Contains(string(encoded), `"entry_counters":[`) {
				t.Fatalf("wire purpose %s: %v", encoded, err)
			}
			if redactCardForViewer(*v, false).Purpose != nil {
				t.Fatal("hidden card leaked entry amounts")
			}
			c.FaceDown, c.FaceDownKind = true, game.FaceDownMorphed
			if game.CardPurposeOf(c).EntryCounters != nil {
				t.Fatal("face-down card exposed entry amounts")
			}
		})
	}
}
