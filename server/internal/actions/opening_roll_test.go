package actions

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand/v2"
	"sort"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opening_roll_test.go — ADR 0121 §1 and §3 at the dispatch layer: the
// allowlist gate, the three verbs, and MintsNoUndo.

// newOpeningRollGame seats four players and opens the opening roll on a
// key whose round 1 ties seats 2 and 3 (7, 4, 8, 8).
func newOpeningRollGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(22, 2048))); err != nil {
		t.Fatalf("StartWithOpeningRoll: %v", err)
	}
	return g
}

// declaredActionTypes reads every `Type` constant out of actions.go, so
// a type added tomorrow is in this list without anybody remembering to
// put it there.
func declaredActionTypes(t *testing.T) []Type {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "actions.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var out []Type
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			if id, ok := vs.Type.(*ast.Ident); !ok || id.Name != "Type" {
				continue
			}
			for _, v := range vs.Values {
				lit, ok := v.(*ast.BasicLit)
				if !ok {
					continue
				}
				s, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatal(err)
				}
				out = append(out, Type(s))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// The gate fails closed: every action type outside the allowlist is
// refused while the roll is open, including any added after this test
// was written. A new type that should be allowed goes in
// openingRollActions; one that should not needs nothing.
func TestEveryActionButTheOpeningRollsIsRefusedWhileItIsOpen(t *testing.T) {
	types := declaredActionTypes(t)
	if len(types) < 40 {
		t.Fatalf("found only %d action types in actions.go — the scanner has stopped working", len(types))
	}
	g := newOpeningRollGame(t)
	p := g.Seats[0].ID
	for _, typ := range types {
		if _, allowed := openingRollActions[typ]; allowed {
			continue
		}
		for _, caller := range []uuid.UUID{p, uuid.Nil} {
			err := Dispatch(g, Action{Type: typ, Player: p, Caller: caller, Params: []byte(`{}`)})
			if !errors.Is(err, game.ErrOpeningRollOpen) {
				t.Errorf("%s from caller %v while the opening roll is open: %v, want ErrOpeningRollOpen", typ, caller, err)
			}
		}
	}
	if err := Dispatch(g, Action{Type: "no_such_action", Player: p, Caller: p}); !errors.Is(err, game.ErrOpeningRollOpen) {
		t.Errorf("an unknown type while the roll is open: %v, want ErrOpeningRollOpen", err)
	}
	if !g.OpeningRollOpen() || len(g.Seats[0].Hand.Cards) != 0 {
		t.Fatal("a refused action changed the game")
	}
	for typ := range openingRollActions {
		known := false
		for _, d := range types {
			known = known || d == typ
		}
		if !known {
			t.Errorf("openingRollActions names %q, which actions.go does not declare", typ)
		}
	}
}

func TestOpeningRollVerbsThroughDispatch(t *testing.T) {
	g := newOpeningRollGame(t)
	seat := func(i int) uuid.UUID { return g.Seats[i].ID }

	// roll_opening is player-scoped: a seat rolls its own die only.
	if err := Dispatch(g, Action{Type: TypeRollOpening, Player: seat(1), Caller: seat(0)}); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 0 rolled for seat 1: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypeRollOpening, Caller: seat(0)}); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("a roll with no player: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypeRollOpening, Player: seat(0), Caller: seat(0)}); err != nil {
		t.Fatal(err)
	}
	if err := Dispatch(g, Action{Type: TypeRollOpening, Player: seat(0), Caller: seat(0)}); !errors.Is(err, game.ErrNotYourRoll) {
		t.Fatalf("a second die: %v", err)
	}
	// The host's button records the caller as the presser.
	if err := Dispatch(g, Action{Type: TypeHostRollRemaining, Caller: seat(0)}); err != nil {
		t.Fatal(err)
	}
	if by := g.OpeningRoll.Rounds[0].Rolls[1].By; by != 0 {
		t.Fatalf("by = %d, want the host's seat 0", by)
	}
	// Round 2 (seats 2 and 3), by the admin.
	if err := Dispatch(g, Action{Type: TypeHostRollRemaining}); err != nil {
		t.Fatal(err)
	}
	chooser := g.OpeningRoll.Chooser
	if chooser != 3 {
		t.Fatalf("chooser = %d, want 3", chooser)
	}

	choose := func(player, caller uuid.UUID, raw string) error {
		return Dispatch(g, Action{Type: TypeChooseStartingPlayer, Player: player, Caller: caller, Params: []byte(raw)})
	}
	if err := choose(seat(3), seat(3), `{}`); !errors.Is(err, ErrMissingParams) {
		t.Fatalf("a choice with no seat: %v, want ErrMissingParams", err)
	}
	if err := choose(seat(3), seat(2), `{"seat":2}`); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Fatalf("seat 2 chose for seat 3: %v", err)
	}
	if err := choose(seat(2), seat(2), `{"seat":2}`); !errors.Is(err, game.ErrNotChooser) {
		t.Fatalf("the loser chose: %v", err)
	}
	// The admin acts for the chooser; seat 0 takes the first turn.
	if err := choose(seat(3), uuid.Nil, `{"seat":0}`); err != nil {
		t.Fatal(err)
	}
	if g.OpeningRollOpen() || g.StartingSeat != 0 || len(g.Seats[0].Hand.Cards) != game.OpeningHandSize {
		t.Fatalf("after the choice: open %v, starting seat %d, hand %d", g.OpeningRollOpen(), g.StartingSeat, len(g.Seats[0].Hand.Cards))
	}
	// The window is closed, so the gate lets the mulligan through.
	if err := Dispatch(g, Action{Type: TypeKeepHand, Player: seat(0), Caller: seat(0)}); err != nil {
		t.Fatalf("keep_hand after the choice: %v", err)
	}
}

func TestConcedeIsAllowedWhileTheOpeningRollIsOpen(t *testing.T) {
	g := newOpeningRollGame(t)
	p := g.Seats[1].ID
	if err := Dispatch(g, Action{Type: TypeConcede, Player: p, Caller: p}); err != nil {
		t.Fatal(err)
	}
	if !g.Seats[1].Eliminated || !g.OpeningRollOpen() {
		t.Fatalf("eliminated %v, roll open %v", g.Seats[1].Eliminated, g.OpeningRollOpen())
	}
}

func TestMintsNoUndo(t *testing.T) {
	open := newOpeningRollGame(t)
	closed := newGame(t)
	always := []Type{TypeRollOpening, TypeHostRollRemaining, TypeChooseStartingPlayer, TypeSetTableSettings, TypeSetUndoLimit}
	for _, typ := range always {
		if !MintsNoUndo(open, typ) || !MintsNoUndo(closed, typ) {
			t.Errorf("%s mints an undo entry", typ)
		}
	}
	for _, typ := range []Type{TypeConcede, TypePassPriority, TypeCastSpell} {
		if !MintsNoUndo(open, typ) {
			t.Errorf("%s mints an undo entry while the opening roll is open", typ)
		}
		if MintsNoUndo(closed, typ) {
			t.Errorf("%s mints no undo entry in an ordinary game", typ)
		}
	}
}
