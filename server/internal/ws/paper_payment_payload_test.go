package ws

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #3855: the paper_payment_required frame carries the shortfall slot by
// slot and the card the refused action was about — the spell of a
// cast_spell, the source of an activate_ability — so the client can
// find the payload to re-send.
func TestPaperPaymentPayload(t *testing.T) {
	pp := &game.PaperPaymentRequiredError{
		Slots:           []game.PaperSlot{{Symbol: "{B}", Options: []string{"B"}}, {Symbol: "{1}", Options: []string{"W", "U", "B", "R", "G", "C"}, Count: 1000}},
		ArtifactSources: true,
	}
	for _, tc := range []struct {
		kind, params string
	}{
		{"cast_spell", `{"instance_id":"abc"}`},
		{"activate_ability", `{"source_card_id":"abc","ability_index":0}`},
	} {
		body := paperPaymentPayload(pp, protocol.ActionPayload{Type: tc.kind, Params: json.RawMessage(tc.params)})
		if body.Code != protocol.CodePaperPaymentRequired || body.CardID != "abc" || body.PaperPayment == nil {
			t.Fatalf("%s: payload = %+v", tc.kind, body)
		}
		// #3904: a counted slot keeps its count.
		want := []protocol.PaperPaymentSlot{{Symbol: "{B}", Options: []string{"B"}}, {Symbol: "{1}", Options: []string{"W", "U", "B", "R", "G", "C"}, Count: 1000}}
		if !reflect.DeepEqual(body.PaperPayment.Slots, want) || !body.PaperPayment.ArtifactSources {
			t.Errorf("%s: prompt = %+v", tc.kind, body.PaperPayment)
		}
	}
}
