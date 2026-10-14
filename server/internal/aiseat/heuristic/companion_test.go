package heuristic

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestCompanionIsPricedAsAKnownCardDrawn is ADR 0172 §7: paying {3} to
// put the companion into hand is worth a card in hand and a little for
// the card's size: above the flat special-action price, below a real
// spell of the same size.
func TestCompanionIsPricedAsAKnownCardDrawn(t *testing.T) {
	me := uuid.New()
	companion := uuid.New()
	var v protocol.GameView
	v.Seats = []protocol.PlayerView{{
		ID: me.String(),
		Companion: &protocol.CardView{
			InstanceID: companion.String(), Name: "Jegantha, the Wellspring", Owner: me.String(),
			ManaCost: "{4}{R/G}", IsCompanion: true,
		},
	}}
	v.Turn.ActiveSeat = 0
	v.Turn.Step = "precombat_main"
	raw, _ := json.Marshal(map[string]any{"card_id": companion.String(), "kind": "companion"})
	move := legal.Move{Type: legal.TypeSpecialAction, Kind: legal.KindSpecialAction, Source: companion, Params: raw}

	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: v})
	got, reason := p.payoffOf(st, move)
	want := st.w.Hand * 1.5 // a card in hand, plus a tenth of one per point of mana value 5
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("companion = %v (%s), want %v", got, reason, want)
	}
	if got <= p.cfg.SpecialActionValue {
		t.Errorf("companion %v is not above the flat special action %v", got, p.cfg.SpecialActionValue)
	}
	if spell := p.cfg.SpellPerMana * 5; got >= spell {
		t.Errorf("companion %v is not below a five-mana spell %v", got, spell)
	}
}
