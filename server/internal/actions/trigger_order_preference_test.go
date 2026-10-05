package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// trigger_order_preference_test.go — #1530 at the dispatch layer.

func TestSetTriggerOrderPreferenceThroughDispatch(t *testing.T) {
	g := newGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	set := func(player, caller uuid.UUID, params string) error {
		return Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: player, Caller: caller, Params: []byte(params)})
	}
	if err := set(b, a, `{"always_ask":true}`); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 0 set seat 1's preference: %v", err)
	}
	if err := set(uuid.Nil, a, `{"always_ask":true}`); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("no player: %v", err)
	}
	if err := set(a, a, ``); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no params: %v", err)
	}
	if err := set(a, a, `{}`); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no always_ask: %v", err)
	}
	if err := set(a, a, `{"always_ask":"yes"}`); err == nil {
		t.Fatal("accepted a string")
	}
	if err := set(a, a, `{"always_ask":true}`); err != nil {
		t.Fatal(err)
	}
	if !g.Seats[0].TriggerOrderAlwaysAsk || g.Seats[1].TriggerOrderAlwaysAsk {
		t.Fatal("preference not set on exactly seat 0")
	}
	if err := set(a, a, `{"always_ask":false}`); err != nil || g.Seats[0].TriggerOrderAlwaysAsk {
		t.Fatalf("turning it off: %v", err)
	}
	// The admin may set any seat's.
	if err := set(b, uuid.Nil, `{"always_ask":true}`); err != nil || !g.Seats[1].TriggerOrderAlwaysAsk {
		t.Fatalf("admin set: %v", err)
	}
}

func TestSetTriggerOrderPreferenceIsASettingNotAPlay(t *testing.T) {
	g := newOpeningRollGame(t)
	if !MintsNoUndo(g, TypeSetTriggerOrderPreference) {
		t.Error("the preference mints an undo entry")
	}
	p := g.Seats[2].ID
	err := Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: p, Caller: p, Params: []byte(`{"always_ask":true}`)})
	if err != nil {
		t.Fatalf("refused during the opening roll: %v", err)
	}
}
