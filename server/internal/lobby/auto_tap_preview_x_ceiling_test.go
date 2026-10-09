package lobby

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_x_ceiling_test.go — #2581. The preview reports a
// spell's printed X ceiling, read now for the caster, beside the price:
// the number CastSpell refuses above, so the X picker can stop at it.
// Open the Way's "X can't be greater than the number of players in the
// game" at a two-seat table is 2.
func TestAutoTapPreviewReportsThePrintedXCeiling(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	way := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Open the Way", TypeLine: "Sorcery", ManaCost: "{X}{G}{G}",
		OracleID: "e07297c7-83bc-4162-bbb1-362bc737efe0",
	}, 1)[0]
	plain := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Plain X", TypeLine: "Sorcery", ManaCost: "{X}{R}",
	}, 1)[0]

	read := func(query string) *int {
		t.Helper()
		resp := f.get("/games/" + f.gameID.String() + "/auto-tap-preview?" + query)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d", resp.StatusCode)
		}
		var body struct {
			XMax *int `json:"x_max"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return body.XMax
	}
	if got := read("card=" + way.String() + "&x=1"); got == nil || *got != 2 {
		t.Errorf("Open the Way at two seats: x_max = %v, want 2", got)
	}
	if got := read("card=" + plain.String() + "&x=1"); got != nil {
		t.Errorf("a card with no ceiling: x_max = %d, want absent", *got)
	}
}
