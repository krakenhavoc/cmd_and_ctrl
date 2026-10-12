package protocol

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulse_view_test.go — #2480: a card exiled that its controller may
// play this turn rides the purpose as `impulse`, and is Priced, so the
// bot drops its flat price for it.
func TestPurposeViewCarriesImpulse(t *testing.T) {
	v := viewOfPurpose(game.Purpose{Impulse: 1})
	if v == nil || *v != (PurposeView{Impulse: 1}) {
		t.Fatalf("Professional Face-Breaker's purpose = %+v, want impulse 1", v)
	}
	if !v.Priced() {
		t.Error("an impulse purpose is not Priced")
	}
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"impulse":1}` {
		t.Errorf("wire = %s", got)
	}
}
