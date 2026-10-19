package actions

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// splice_payments_test.go — #4055, ADR 0202's 2026-10-19 amendment: a
// cast_spell carries what each revealed splice card's cost that is not mana
// is paid with, entry k paying splices[k].

func TestSplicePaymentWireMapsEveryField(t *testing.T) {
	ids := make([]uuid.UUID, 6)
	for i := range ids {
		ids[i] = uuid.New()
	}
	var w splicePaymentWire
	raw := params(t, map[string]any{
		"sacrifice_ids": []string{ids[0].String(), ids[1].String()},
		"exile_ids":     []string{ids[2].String()},
		"tap_ids":       []string{ids[3].String()},
		"return_ids":    []string{ids[4].String()},
		"opponent":      ids[5].String(),
	})
	if err := json.Unmarshal(raw, &w); err != nil {
		t.Fatal(err)
	}
	pay, err := w.toPayment()
	if err != nil {
		t.Fatalf("toPayment: %v", err)
	}
	if len(pay.SacrificeIDs) != 2 || pay.SacrificeIDs[0] != ids[0] || pay.SacrificeIDs[1] != ids[1] {
		t.Errorf("sacrifice ids %v", pay.SacrificeIDs)
	}
	if len(pay.ExileIDs) != 1 || pay.ExileIDs[0] != ids[2] {
		t.Errorf("exile ids %v", pay.ExileIDs)
	}
	if len(pay.TapIDs) != 1 || pay.TapIDs[0] != ids[3] {
		t.Errorf("tap ids %v", pay.TapIDs)
	}
	if len(pay.ReturnIDs) != 1 || pay.ReturnIDs[0] != ids[4] {
		t.Errorf("return ids %v", pay.ReturnIDs)
	}
	if pay.Opponent != ids[5] {
		t.Errorf("opponent %v", pay.Opponent)
	}
}

func TestSplicePaymentWireRefusesAMalformedID(t *testing.T) {
	for _, field := range []string{"sacrifice_ids", "exile_ids", "tap_ids", "return_ids"} {
		var w splicePaymentWire
		if err := json.Unmarshal(params(t, map[string]any{field: []string{"nope"}}), &w); err != nil {
			t.Fatal(err)
		}
		if _, err := w.toPayment(); err == nil {
			t.Errorf("%s: a malformed id was accepted", field)
		}
	}
	var w splicePaymentWire
	if err := json.Unmarshal(params(t, map[string]any{"opponent": "nope"}), &w); err != nil {
		t.Fatal(err)
	}
	if _, err := w.toPayment(); err == nil {
		t.Error("a malformed opponent was accepted")
	}
}

// A malformed entry refuses the whole cast before the engine hears of it.
func TestCastSpellRefusesAMalformedSplicePayment(t *testing.T) {
	g := newGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cast, err := Decode(string(TypeCastSpell), me.ID.String(), params(t, map[string]any{
		"instance_id":     uuid.New().String(),
		"splices":         []string{uuid.New().String()},
		"splice_payments": []map[string]any{{"return_ids": []string{"nope"}}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	cast.Caller = me.ID
	if err := Dispatch(g, cast); err == nil {
		t.Error("a malformed splice payment was accepted")
	}
}
