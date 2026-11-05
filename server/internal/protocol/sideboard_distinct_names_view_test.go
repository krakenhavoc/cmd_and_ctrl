package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sideboard_distinct_names_view_test.go — #4962: with a sideboard seated,
// an outside pick "with different names" (Turtles Forever's outside half)
// is a choose_cards over the sideboard, and it says its names must differ
// as the database prompt does (#4947), so the picker never lets a second
// card with a picked name be clicked. A pick whose names may repeat
// (Research) does not say so. Like the bounds, the flag is the chooser's.
func TestASideboardPickSaysWhenItsNamesMustDiffer(t *testing.T) {
	for _, tc := range []struct {
		name string
		pick game.OutsidePick
		want bool
	}{
		{"different names", testProtocolDistinctPick, true},
		{"names may repeat", testProtocolRepeatPick, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := sideboardViewGame(t)
			// Seat 0's deck seated a sideboard: its pool is that.
			chooser := g.Seats[0]
			g.WithWriteLock(func() {
				if !g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
					Player: chooser.ID, Question: "choose up to two cards", Max: 2,
					Keyed: tc.pick,
				}) {
					t.Fatal("no prompt")
				}
			})
			own := ViewOfGameFor(g, chooser.ID.String())
			if len(own.PendingChoices) != 1 {
				t.Fatalf("%d pending choices", len(own.PendingChoices))
			}
			pc := own.PendingChoices[0]
			if pc.Kind != string(game.PendingChoiceChooseCards) || pc.ChooseMax != 2 || len(pc.Options) != 2 {
				t.Fatalf("chooser's prompt %q max %d, %d options", pc.Kind, pc.ChooseMax, len(pc.Options))
			}
			if pc.DistinctNames != tc.want {
				t.Fatalf("distinct_names = %v, want %v", pc.DistinctNames, tc.want)
			}
			raw, _ := json.Marshal(pc)
			if got := strings.Contains(string(raw), `"distinct_names":true`); got != tc.want {
				t.Fatalf("the chooser's wire carries distinct_names: %v, want %v (%s)", got, tc.want, raw)
			}
			for _, viewer := range []string{g.Seats[1].ID.String(), SpectatorViewerID, ""} {
				v := ViewOfGameFor(g, viewer)
				if len(v.PendingChoices) != 1 {
					t.Fatalf("viewer %q: %d pending choices", viewer, len(v.PendingChoices))
				}
				if v.PendingChoices[0].DistinctNames {
					t.Errorf("viewer %q sees the chooser's distinct_names", viewer)
				}
				raw, _ := json.Marshal(v)
				if strings.Contains(string(raw), "distinct_names") {
					t.Errorf("viewer %q: the wire carries distinct_names", viewer)
				}
			}
		})
	}
}
