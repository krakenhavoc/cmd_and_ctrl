package heuristic

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0189 decision 4: a manifested Den Protector offers its megamorph
// cost and its mana cost for the same {1}{G}; the bot takes the one
// that adds the counter.
func TestBotTakesTheMegamorphRoute(t *testing.T) {
	me := uuid.New()
	card := uuid.New()
	var v protocol.GameView
	v.Seats = []protocol.PlayerView{{ID: me.String()}}
	v.Turn.ActiveSeat = 0
	v.Turn.Step = "precombat_main"

	move := func(route, label string) legal.Move {
		raw, _ := json.Marshal(map[string]any{"card_id": card.String(), "kind": "turn_face_up", "route": route})
		return legal.Move{Type: legal.TypeSpecialAction, Kind: legal.KindSpecialAction, Source: card, Label: label, Params: raw}
	}
	megamorph := move("morph", "Turn face up {1}{G} (megamorph, +1/+1 counter)")
	manaCost := move("mana_cost", "Turn face up {1}{G} (mana cost)")
	plainMorph := move("morph", "Turn face up {1}{G} (morph)")

	p := New()
	st := p.newState(aiseat.Input{Seat: me, View: v})
	mega, _ := p.payoffOf(st, megamorph)
	mana, _ := p.payoffOf(st, manaCost)
	plain, _ := p.payoffOf(st, plainMorph)
	if mega <= mana {
		t.Errorf("megamorph route %v does not beat the mana-cost route %v", mega, mana)
	}
	if mana != p.cfg.SpecialActionValue || plain != p.cfg.SpecialActionValue {
		t.Errorf("mana cost = %v, plain morph = %v, want the flat %v", mana, plain, p.cfg.SpecialActionValue)
	}
}
