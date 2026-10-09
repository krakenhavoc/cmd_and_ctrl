package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_order_preference_test.go — #1530 and #1968 at the dispatch
// layer: the three-way trigger_order, and the old always_ask boolean.

func TestSetTriggerOrderPreferenceThroughDispatch(t *testing.T) {
	g := newGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	set := func(player, caller uuid.UUID, params string) error {
		return Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: player, Caller: caller, Params: []byte(params)})
	}
	if err := set(b, a, `{"trigger_order":"never"}`); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 0 set seat 1's preference: %v", err)
	}
	if err := set(uuid.Nil, a, `{"trigger_order":"never"}`); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("no player: %v", err)
	}
	if err := set(a, a, ``); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no params: %v", err)
	}
	if err := set(a, a, `{}`); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no trigger_order: %v", err)
	}
	for _, bad := range []string{`{"trigger_order":"sometimes"}`, `{"trigger_order":true}`, `{"always_ask":"yes"}`} {
		if err := set(a, a, bad); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	for wire, want := range map[string]game.TriggerOrderMode{
		"never":           game.TriggerOrderNever,
		"always":          game.TriggerOrderAlways,
		"when_it_matters": game.TriggerOrderWhenItMatters,
	} {
		if err := set(a, a, `{"trigger_order":"`+wire+`"}`); err != nil {
			t.Fatalf("%s: %v", wire, err)
		}
		if g.Seats[0].TriggerOrder != want || g.Seats[1].TriggerOrder != game.TriggerOrderWhenItMatters {
			t.Fatalf("%s: seats are %q / %q", wire, g.Seats[0].TriggerOrder, g.Seats[1].TriggerOrder)
		}
	}
	// The admin may set any seat's.
	if err := set(b, uuid.Nil, `{"trigger_order":"always"}`); err != nil || g.Seats[1].TriggerOrder != game.TriggerOrderAlways {
		t.Fatalf("admin set: %v", err)
	}
}

// A client from before #1968 sends {always_ask}: true is always, false
// is when_it_matters. trigger_order wins when both are sent.
func TestSetTriggerOrderPreferenceReadsTheOldBoolean(t *testing.T) {
	g := newGame(t)
	a := g.Seats[0].ID
	set := func(params string) error {
		return Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: a, Caller: a, Params: []byte(params)})
	}
	if err := set(`{"always_ask":true}`); err != nil || g.Seats[0].TriggerOrder != game.TriggerOrderAlways {
		t.Fatalf("always_ask true: %v, %q", err, g.Seats[0].TriggerOrder)
	}
	if err := set(`{"always_ask":false}`); err != nil || g.Seats[0].TriggerOrder != game.TriggerOrderWhenItMatters {
		t.Fatalf("always_ask false: %v, %q", err, g.Seats[0].TriggerOrder)
	}
	if err := set(`{"always_ask":true,"trigger_order":"never"}`); err != nil || g.Seats[0].TriggerOrder != game.TriggerOrderNever {
		t.Fatalf("both: %v, %q", err, g.Seats[0].TriggerOrder)
	}
}

func TestSetTriggerOrderPreferenceIsASettingNotAPlay(t *testing.T) {
	g := newOpeningRollGame(t)
	if !MintsNoUndo(g, TypeSetTriggerOrderPreference) {
		t.Error("the preference mints an undo entry")
	}
	p := g.Seats[2].ID
	err := Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: p, Caller: p, Params: []byte(`{"trigger_order":"never"}`)})
	if err != nil {
		t.Fatalf("refused during the opening roll: %v", err)
	}
}
