package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #5781, CR 116.2c: a licid in its Aura form offers end_effect to the
// player whose ability made it an Aura, whoever controls it now, and to
// nobody else.
func TestEnumeratorOffersEndEffectToTheLicidsActivator(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	licid := game.NewCard("Licid", other.ID)
	licid.TypeLine = "Creature — Licid"
	licid.Controller = other.ID
	licid.ReturnedAura = true
	licid.AuraFormEndBy = seat.ID
	licid.AuraFormEndCost = "{0}"
	g.WithWriteLock(func() { g.Battlefield.PushTop(licid) })
	endMoves := func(who uuid.UUID) []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, who) {
			if m.Kind == legal.KindSpecialAction && m.Source == licid.InstanceID {
				out = append(out, m)
			}
		}
		return out
	}
	got := endMoves(seat.ID)
	if len(got) != 1 {
		t.Fatalf("end_effect moves for the activator = %v, want 1", labels(got))
	}
	var p struct {
		Kind   string `json:"kind"`
		CardID string `json:"card_id"`
	}
	if err := json.Unmarshal(got[0].Params, &p); err != nil {
		t.Fatal(err)
	}
	if p.Kind != "end_effect" || p.CardID != licid.InstanceID.String() {
		t.Errorf("params = %+v, want end_effect on the licid", p)
	}
	if other := endMoves(other.ID); len(other) != 0 {
		t.Errorf("the licid's controller, who did not activate it, is offered %v", labels(other))
	}
}
