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

func TestDispatchPassPriorityRotates(t *testing.T) {
	// 2-player game: one PassPriority rotates priority to seat 1
	// without advancing the step (the next would wrap and advance).
	g := newGame(t)
	beforeStep := g.Turn.Step
	beforeHolder := g.Turn.PriorityHolder
	a, _ := Decode(string(TypePassPriority), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if g.Turn.Step != beforeStep {
		t.Errorf("step changed unexpectedly: %q -> %q", beforeStep, g.Turn.Step)
	}
	if g.Turn.PriorityHolder == beforeHolder {
		t.Errorf("priority did not rotate (still %d)", g.Turn.PriorityHolder)
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

func TestDispatchPassPriorityRejectsNonHolder(t *testing.T) {
	g := newGame(t)
	holder := g.Seats[g.Turn.PriorityHolder].ID
	other := g.Seats[(g.Turn.PriorityHolder+1)%len(g.Seats)].ID

	a, _ := Decode(string(TypePassPriority), "", nil)
	a.Caller = other
	if err := Dispatch(g, a); !errors.Is(err, ErrNotPriorityHolder) {
		t.Errorf("non-holder dispatch: got %v, want ErrNotPriorityHolder", err)
	}
	// State must not have moved.
	if g.Seats[g.Turn.PriorityHolder].ID != holder {
		t.Errorf("priority holder changed despite rejected dispatch")
	}

	// The legitimate holder still succeeds.
	a.Caller = holder
	if err := Dispatch(g, a); err != nil {
		t.Errorf("holder dispatch: got %v, want nil", err)
	}
}

func TestDispatchPassTurnRejectsNonActivePlayer(t *testing.T) {
	g := newGame(t)
	active := g.Seats[g.Turn.ActiveSeat].ID
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID

	a, _ := Decode(string(TypePassTurn), "", nil)
	a.Caller = other
	if err := Dispatch(g, a); !errors.Is(err, ErrNotActivePlayer) {
		t.Errorf("non-active dispatch: got %v, want ErrNotActivePlayer", err)
	}
	if g.Seats[g.Turn.ActiveSeat].ID != active {
		t.Errorf("active seat changed despite rejected dispatch")
	}

	// The active player still succeeds.
	a.Caller = active
	if err := Dispatch(g, a); err != nil {
		t.Errorf("active-player dispatch: got %v, want nil", err)
	}
}

func TestDispatchRejectsCrossSeatPlayerScopedAction(t *testing.T) {
	g := newGame(t)
	caller := g.Seats[0].ID
	target := g.Seats[1].ID

	// change_life on someone else's seat is the prototypical case.
	a, _ := Decode(string(TypeChangeLife), target.String(), params(t, struct {
		Delta int `json:"delta"`
	}{Delta: -40}))
	a.Caller = caller
	if err := Dispatch(g, a); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("cross-seat change_life: got %v, want ErrPlayerCallerMismatch", err)
	}
	if g.Seats[1].Life != 40 {
		t.Errorf("target life changed despite rejected dispatch: %d", g.Seats[1].Life)
	}

	// Self-targeting still works.
	a, _ = Decode(string(TypeChangeLife), caller.String(), params(t, struct {
		Delta int `json:"delta"`
	}{Delta: -3}))
	a.Caller = caller
	if err := Dispatch(g, a); err != nil {
		t.Errorf("self-targeted change_life: got %v, want nil", err)
	}
	if g.Seats[0].Life != 37 {
		t.Errorf("self life: got %d, want 37", g.Seats[0].Life)
	}
}

func TestDispatchPassPriorityAllowsAdminCaller(t *testing.T) {
	// Admin / spectator (Caller == uuid.Nil) bypasses the gate so a
	// trusted moderator can advance the game on a player's behalf.
	g := newGame(t)
	a, _ := Decode(string(TypePassPriority), "", nil)
	if err := Dispatch(g, a); err != nil {
		t.Errorf("admin dispatch: got %v, want nil", err)
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

func TestDispatchSetBattlefieldPosition(t *testing.T) {
	g := newGame(t)
	p := g.Seats[0]
	_ = g.DrawCard(p.ID)
	card, _ := p.Hand.Top()
	_ = g.PlayCard(p.ID, card.InstanceID)

	a, _ := Decode(
		string(TypeSetBattlefieldPosition),
		"",
		params(t, map[string]any{
			"instance_id": card.InstanceID.String(),
			"x":           0.4,
			"y":           0.6,
		}),
	)
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == card.InstanceID {
			if c.BattleX != 0.4 || c.BattleY != 0.6 {
				t.Errorf("position: got (%v, %v), want (0.4, 0.6)", c.BattleX, c.BattleY)
			}
		}
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
