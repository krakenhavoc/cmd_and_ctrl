package protocol

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lands_untapped_view_test.go — ADR 0136 §2: a ramp spell's untapped
// lands ride the purpose as `lands_untapped`, and stay off the wire when
// the lands enter tapped.
func TestPurposeViewCarriesLandsUntapped(t *testing.T) {
	v := viewOfPurpose(game.Purpose{Lands: 2, LandsUntapped: 2})
	if v == nil || *v != (PurposeView{Lands: 2, LandsUntapped: 2}) {
		t.Fatalf("Harrow's purpose = %+v, want lands 2, lands_untapped 2", v)
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"lands":2,"lands_untapped":2}` {
		t.Errorf("wire = %s", got)
	}
	b, _ = json.Marshal(viewOfPurpose(game.Purpose{Lands: 1}))
	if got := string(b); got != `{"lands":1}` {
		t.Errorf("Rampant Growth's wire = %s, want no lands_untapped", got)
	}
}
