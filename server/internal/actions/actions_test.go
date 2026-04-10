package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func newGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 2 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return g
}

func params(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	return raw
}

func TestDispatchDrawCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]

	a, err := Decode(string(TypeDrawCard), p.ID.String(), nil)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p.Hand.Size() != 1 {
		t.Errorf("hand size: got %d, want 1", p.Hand.Size())
	}
}

func TestDispatchPlayCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	if err := g.DrawCard(p.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	card, _ := p.Hand.Top()

	a, _ := Decode(
		string(TypePlayCard),
		p.ID.String(),
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Battlefield.Contains(card.InstanceID) {
		t.Error("card not on battlefield after play")
	}
}

func TestDispatchMoveCard(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeMoveCard),
		"",
		params(t, map[string]any{
			"src":         map[string]string{"kind": "battlefield"},
			"dst":         map[string]string{"kind": "exile"},
			"instance_id": card.InstanceID.String(),
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("card not in exile after move_card")
	}
}

func TestDispatchTapAndUntap(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	tapAction, _ := Decode(
		string(TypeTap),
		"",
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, tapAction); err != nil {
		t.Fatalf("tap Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && !c.Tapped {
			t.Error("card should be tapped after tap action")
		}
	}

	untapAction, _ := Decode(
		string(TypeUntap),
		"",
		params(t, map[string]string{"instance_id": card.InstanceID.String()}),
	)
	if err := Dispatch(g, untapAction); err != nil {
		t.Fatalf("untap Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Tapped {
			t.Error("card should be untapped after untap action")
		}
	}
}

func TestDispatchUntapAllRequiresPlayer(t *testing.T) {
	g := newGame(t)
	a, _ := Decode(string(TypeUntapAll), "", nil)
	if err := Dispatch(g, a); !errors.Is(err, ErrInvalidPlayer) {
		t.Errorf("untap_all without player: got %v, want ErrInvalidPlayer", err)
	}
}

func TestDispatchPassPriorityAdvancesStep(t *testing.T) {
	g := newGame(t)
	before := g.Turn.Step
	a, _ := Decode(string(TypePassPriority), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Turn.Step == before {
		t.Error("pass_priority did not advance")
	}
}

func TestDispatchPassTurnWrapsSeat(t *testing.T) {
	g := newGame(t)
	a, _ := Decode(string(TypePassTurn), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("seat after pass_turn: got %d, want 1", g.Turn.ActiveSeat)
	}
}

func TestDispatchMulligan(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	for range 7 {
		_ = g.DrawCard(p.ID)
	}
	a, _ := Decode(
		string(TypeMulligan),
		p.ID.String(),
		params(t, map[string]int{"hand_size": 6}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p.Hand.Size() != 6 {
		t.Errorf("hand size: got %d, want 6", p.Hand.Size())
	}
}

func TestDispatchShuffleLibrary(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	before := append([]game.Card(nil), p.Library.Cards...)
	a, _ := Decode(string(TypeShuffleLibrary), p.ID.String(), nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	diff := false
	for i := range before {
		if before[i].InstanceID != p.Library.Cards[i].InstanceID {
			diff = true
			break
		}
	}
	if !diff {
		t.Error("library unchanged after shuffle")
	}
}

func TestDispatchChangeLife(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	a, _ := Decode(
		string(TypeChangeLife),
		p.ID.String(),
		params(t, map[string]int{"delta": -3}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p.Life != 37 {
		t.Errorf("life: got %d, want 37", p.Life)
	}
}

func TestDispatchAddCounter(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeAddCounter),
		"",
		params(t, map[string]any{
			"instance_id": card.InstanceID.String(),
			"name":        "+1/+1",
			"delta":       2,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID && c.Counters["+1/+1"] != 2 {
			t.Errorf("counter: got %d, want 2", c.Counters["+1/+1"])
		}
	}
}

func TestDispatchSetCommanderDamage(t *testing.T) {
	g := newGame(t)
	p0, p1 := g.Seats[0], g.Seats[1]
	a, _ := Decode(
		string(TypeSetCommanderDamage),
		"",
		params(t, map[string]any{
			"from":   p0.ID.String(),
			"to":     p1.ID.String(),
			"amount": 12,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if p1.CommanderDamage[p0.ID] != 12 {
		t.Errorf("cmdr damage: got %d, want 12", p1.CommanderDamage[p0.ID])
	}
}

func TestDispatchUnknownType(t *testing.T) {
	g := newGame(t)
	a, _ := Decode("explode", "", nil)
	err := Dispatch(g, a)
	if !errors.Is(err, ErrUnknownType) {
		t.Errorf("unknown type: got %v, want ErrUnknownType", err)
	}
}

func TestDecodeMalformedPlayer(t *testing.T) {
	_, err := Decode(string(TypeDrawCard), "not-a-uuid", nil)
	if !errors.Is(err, ErrInvalidPlayer) {
		t.Errorf("bad player uuid: got %v, want ErrInvalidPlayer", err)
	}
}

func TestDispatchMalformedParams(t *testing.T) {
	g := newGame(t)
	a := Action{
		Type:   TypePlayCard,
		Player: g.Seats[0].ID,
		Params: json.RawMessage(`{ not json`),
	}
	if err := Dispatch(g, a); err == nil {
		t.Error("expected error on malformed params")
	}
}
