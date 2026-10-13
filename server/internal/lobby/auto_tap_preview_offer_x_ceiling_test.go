package lobby

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_offer_x_ceiling_test.go — #3225. The preview reads the
// X ceiling of the cast it is pricing, so the alternative cost the
// player claimed before the X picker opened bounds the picker: Silver
// Scrutiny's flash offer ("as though it had flash if X is 3 or less",
// CR 601.3c) reports x_max 3, the cast for the printed cost reports none.
func TestAutoTapPreviewReportsTheClaimedOffersXCeiling(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	scrutiny := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Silver Scrutiny", TypeLine: "Sorcery", ManaCost: "{X}{U}{U}",
		OracleID: "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e",
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
	if got := read("card=" + scrutiny.String() + "&x=1&alternative_cost=flash-x-at-most"); got == nil || *got != 3 {
		t.Errorf("the flash offer: x_max = %v, want 3", got)
	}
	if got := read("card=" + scrutiny.String() + "&x=1"); got != nil {
		t.Errorf("the printed cost: x_max = %d, want absent", *got)
	}
}
