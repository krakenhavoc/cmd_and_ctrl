package heuristic

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestUnlockIsPricedLikeCastingTheDoor is ADR 0103 owner decision 5:
// an unlock scores like casting a spell of the unlocked door's mana
// value, so the dear door is worth more than the cheap one, and an
// unimplemented Room falls back to the flat special-action value.
func TestUnlockIsPricedLikeCastingTheDoor(t *testing.T) {
	me := uuid.New()
	room := uuid.New()
	view := func(unimplemented bool) protocol.GameView {
		var v protocol.GameView
		v.Seats = []protocol.PlayerView{{ID: me.String()}}
		v.Turn.ActiveSeat = 0
		v.Turn.Step = "precombat_main"
		v.Battlefield.Cards = []protocol.CardView{{
			InstanceID: room.String(), Controller: me.String(), Owner: me.String(),
			TypeLine: "Enchantment — Room", Layout: "split", Unimplemented: unimplemented,
			Faces: []protocol.CardFaceView{
				{Name: "Cheap Door", ManaCost: "{1}{B}"},
				{Name: "Dear Door", ManaCost: "{6}{B}{B}"},
			},
		}}
		return v
	}
	move := func(door string) legal.Move {
		raw, _ := json.Marshal(map[string]any{"card_id": room.String(), "kind": "unlock", "door": door})
		return legal.Move{Type: legal.TypeSpecialAction, Kind: legal.KindSpecialAction, Source: room, Params: raw}
	}
	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: view(false)})
	cheap, _ := p.payoffOf(st, move("left"))
	dear, reason := p.payoffOf(st, move("right"))
	if want := p.cfg.SpellPerMana * 8; dear != want {
		t.Errorf("right door = %v (%s), want SpellPerMana × 8 = %v", dear, reason, want)
	}
	if cheap >= dear {
		t.Errorf("cheap door %v scored at least the dear one %v", cheap, dear)
	}
	st = p.newState(aiseat.Input{Seat: me, View: view(true)})
	if v, _ := p.payoffOf(st, move("right")); v != p.cfg.SpecialActionValue {
		t.Errorf("an unimplemented Room's unlock = %v, want the flat %v", v, p.cfg.SpecialActionValue)
	}
}
