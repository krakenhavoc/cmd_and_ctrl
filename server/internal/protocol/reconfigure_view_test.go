package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestReconfigureFlagIsPublicOnBothRowsAndOmittedElsewhere(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: "Test Equipment", TypeLine: "Artifact — Equipment",
			Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true},
			ActivatedAbilities: []game.ActivatedAbilityShape{
				{Label: "attach", Reconfigure: true},
				{Label: "unattach", Reconfigure: true},
				{Label: "equip", Equip: true},
			},
		})
	})
	g.BumpLayerVersionForTest()
	v := ViewOfGame(g)
	for _, viewer := range []string{me.ID.String(), g.Seats[1].ID.String(), SpectatorViewerID} {
		fv := FilterViewFor(v, viewer)
		c := findCardView(t, fv.Battlefield, id)
		if c == nil || len(c.ActivatedAbilities) != 3 {
			t.Fatalf("viewer %q: expected three ability rows on %+v", viewer, c)
		}
		rows := c.ActivatedAbilities
		if !rows[0].Reconfigure || !rows[1].Reconfigure || rows[2].Reconfigure || rows[0].Equip || rows[1].Equip {
			t.Errorf("viewer %q: wrong flags: %+v", viewer, rows)
		}
		raw, err := json.Marshal(rows)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(raw), `"reconfigure":true`) != 2 || strings.Contains(string(raw), `"reconfigure":false`) {
			t.Errorf("wire flags: %s", raw)
		}
	}
}
