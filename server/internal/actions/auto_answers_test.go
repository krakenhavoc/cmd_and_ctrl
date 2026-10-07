package actions

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_answers_test.go — ADR 0127 §3 at the dispatch layer:
// set_auto_answers is a seat's own setting, replaced whole.

func TestSetAutoAnswersThroughDispatch(t *testing.T) {
	g := newGame(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	set := func(player, caller uuid.UUID, params string) error {
		return Dispatch(g, Action{Type: TypeSetAutoAnswers, Player: player, Caller: caller, Params: []byte(params)})
	}
	one := `{"rules":[{"key":"k1","answer":"always"}]}`
	if err := set(b, a, one); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 0 set seat 1's rules: %v", err)
	}
	if err := set(uuid.Nil, a, one); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("no player: %v", err)
	}
	if err := set(a, a, ``); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no params: %v", err)
	}
	if err := set(a, a, `{}`); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("no rules: %v", err)
	}
	if err := set(a, a, `{"rules":[{"key":"k1","answer":"sometimes"}]}`); err == nil {
		t.Fatal("accepted a bad answer")
	}
	long := strings.Repeat("k", game.MaxAutoAnswerKeyLen+1)
	if err := set(a, a, `{"rules":[{"key":"`+long+`","answer":"never"}]}`); err == nil {
		t.Fatal("accepted an over-long key")
	}
	if err := set(a, a, `{"rules":[{"key":"k1","answer":"never"},{"key":"k1","answer":"always"}]}`); err == nil {
		t.Fatal("accepted a key named twice")
	}
	var many []string
	for i := 0; i <= game.MaxAutoAnswerRules; i++ {
		many = append(many, fmt.Sprintf(`{"key":"k%d","answer":"never"}`, i))
	}
	if err := set(a, a, `{"rules":[`+strings.Join(many, ",")+`]}`); !errors.Is(err, game.ErrTooManyAutoAnswers) {
		t.Fatalf("101 rules: %v", err)
	}
	if err := set(a, a, `{"rules":[`+strings.Join(many[:game.MaxAutoAnswerRules], ",")+`]}`); err != nil {
		t.Fatalf("100 rules: %v", err)
	}
	if err := set(a, a, one); err != nil {
		t.Fatal(err)
	}
	if got := g.Seats[0].AutoAnswers; len(got) != 1 || got["k1"] != game.AutoAnswerAlways {
		t.Fatalf("rules = %v, want exactly k1: always (the list replaces)", got)
	}
	if len(g.Seats[1].AutoAnswers) != 0 {
		t.Fatal("seat 1 got seat 0's rules")
	}
	if err := set(a, a, `{"rules":[]}`); err != nil || len(g.Seats[0].AutoAnswers) != 0 {
		t.Fatalf("an empty list did not clear: %v %v", err, g.Seats[0].AutoAnswers)
	}
	// The admin may set any seat's.
	if err := set(b, uuid.Nil, one); err != nil || g.Seats[1].AutoAnswers["k1"] != game.AutoAnswerAlways {
		t.Fatalf("admin set: %v", err)
	}
}

func TestSetAutoAnswersIsASettingNotAPlay(t *testing.T) {
	g := newOpeningRollGame(t)
	if !MintsNoUndo(g, TypeSetAutoAnswers) {
		t.Error("set_auto_answers mints an undo entry")
	}
	p := g.Seats[2].ID
	err := Dispatch(g, Action{Type: TypeSetAutoAnswers, Player: p, Caller: p, Params: []byte(`{"rules":[{"key":"k","answer":"never"}]}`)})
	if err != nil {
		t.Fatalf("refused during the opening roll: %v", err)
	}
}
