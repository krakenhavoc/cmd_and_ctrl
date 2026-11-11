package ws

// match_test.go — ADR 0222 §3 at the room: between two games of a
// match only sideboarding's actions are accepted, a refused deck
// changes nothing, and no undo reaches across the end of a game or the
// start of the next.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func newMatchRoom(t *testing.T) (*Room, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := game.NewGame()
	std, three := game.FormatStandard, 3
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Format: &std, BestOf: &three}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Alice", "Bob"} {
		deck := make([]game.Card, 0, 75)
		for i := 0; i < 60; i++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("%s Main %d", name, i+1), uuid.Nil))
		}
		for i := 0; i < 15; i++ {
			c := game.NewCard(fmt.Sprintf("%s Side %d", name, i+1), uuid.Nil)
			c.IsSideboard = true
			deck = append(deck, c)
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 13))); err != nil {
		t.Fatal(err)
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	return room, g.Seats[0].ID, g.Seats[1].ID
}

// send dispatches one action the way the hub does: through Apply, or
// ApplyExternalAction for an action that mints no undo entry.
func send(room *Room, player uuid.UUID, typ actions.Type, params any) (protocol.GameView, error) {
	a := actions.Action{Type: typ, Player: player, Caller: player}
	if params != nil {
		raw, _ := json.Marshal(params)
		a.Params = raw
	}
	fn := func() error { return actions.Dispatch(room.Game, a) }
	if actions.MintsNoUndo(room.Game, typ) {
		v, _, err := room.ApplyExternalAction(ActionTag{Actor: player}, fn)
		return v, err
	}
	v, _, err := room.Apply(player, fn)
	return v, err
}

func snap(t *testing.T, room *Room) protocol.GameView {
	t.Helper()
	v, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestMatchActionsAtTheRoom(t *testing.T) {
	room, alice, bob := newMatchRoom(t)
	g := room.Game

	// Sideboarding's verbs are refused while a game is in play.
	if _, err := send(room, alice, actions.TypeSideboardReady, map[string]bool{"ready": true}); !errors.Is(err, game.ErrNotSideboarding) {
		t.Errorf("sideboard_ready during a game: %v, want ErrNotSideboarding", err)
	}

	for _, p := range []uuid.UUID{alice, bob} {
		if _, err := send(room, p, actions.TypeKeepHand, nil); err != nil {
			t.Fatalf("keep_hand: %v", err)
		}
	}
	if undoDepth(room) == 0 {
		t.Fatal("keeping a hand left no undo entry; the test needs one to see dropped")
	}

	// Bob concedes game 1: the game's end is final.
	view, err := send(room, bob, actions.TypeConcede, nil)
	if err != nil {
		t.Fatalf("concede: %v", err)
	}
	if !g.SideboardingOpen() || view.State != "active" || view.Match == nil || view.Match.Sideboarding == nil {
		t.Fatalf("after game 1: state %q, match %+v", view.State, view.Match)
	}
	if n := undoDepth(room); n != 0 {
		t.Errorf("%d undo entries reach back into game 1", n)
	}
	if _, _, err := room.Undo(alice); !errors.Is(err, ErrNothingToUndo) {
		t.Errorf("undo between games: %v, want ErrNothingToUndo", err)
	}

	// Nothing but sideboarding happens between games.
	for _, typ := range []actions.Type{actions.TypeConcede, actions.TypeDrawCard, actions.TypePassPriority, actions.TypeKeepHand} {
		if _, err := send(room, alice, typ, nil); !errors.Is(err, game.ErrSideboardingOpen) {
			t.Errorf("%s between games: %v, want ErrSideboardingOpen", typ, err)
		}
	}

	// Bob's frame has his pool; a deck one card short is refused and
	// changes nothing, and a one-for-one swap is taken.
	frame := protocol.FilterViewFor(snap(t, room), bob.String())
	pools := frame.Match.Sideboarding.Pools
	if len(pools) != 1 || len(pools[0].Deck) != 60 || len(pools[0].Sideboard) != 15 {
		t.Fatalf("Bob's frame carries %d pools", len(pools))
	}
	main := make([]string, 0, 60)
	for _, c := range pools[0].Deck[1:] {
		main = append(main, c.InstanceID)
	}
	if _, err := send(room, bob, actions.TypeSetMatchDeck, map[string][]string{"main": main}); !errors.Is(err, game.ErrSideboardIllegal) {
		t.Fatalf("a 59-card deck: %v, want ErrSideboardIllegal", err)
	}
	if code, msg := classifyActionError(fmt.Errorf("%w: the deck has 59 cards; a Standard deck needs at least 60", game.ErrSideboardIllegal)); code != protocol.CodeBadRequest || msg == "" {
		t.Errorf("a refused deck is %q %q on the wire, want a bad_request with the reason", code, msg)
	}
	after := protocol.FilterViewFor(snap(t, room), bob.String()).Match.Sideboarding.Pools[0]
	if len(after.Deck) != 60 || after.Deck[0].InstanceID != pools[0].Deck[0].InstanceID {
		t.Fatal("a refused deck changed the pool")
	}
	in := pools[0].Sideboard[0]
	if _, err := send(room, bob, actions.TypeSetMatchDeck, map[string][]string{"main": append(main, in.InstanceID)}); err != nil {
		t.Fatalf("a one-for-one swap: %v", err)
	}
	// Alice cannot choose; Bob, who lost, can.
	if _, err := send(room, alice, actions.TypeChoosePlayFirst, map[string]bool{"play_first": true}); !errors.Is(err, game.ErrNotMatchChooser) {
		t.Errorf("the winner choosing who plays first: %v", err)
	}
	if _, err := send(room, bob, actions.TypeChoosePlayFirst, map[string]bool{"play_first": false}); err != nil {
		t.Fatalf("choose_play_first: %v", err)
	}
	// One seat may not act for the other.
	other := actions.Action{Type: actions.TypeSideboardReady, Player: alice, Caller: bob, Params: json.RawMessage(`{"ready":true}`)}
	if err := actions.Dispatch(g, other); !errors.Is(err, actions.ErrPlayerCallerMismatch) {
		t.Errorf("readying up for the other seat: %v, want ErrPlayerCallerMismatch", err)
	}

	for _, p := range []uuid.UUID{alice, bob} {
		if view, err = send(room, p, actions.TypeSideboardReady, map[string]bool{"ready": true}); err != nil {
			t.Fatalf("sideboard_ready: %v", err)
		}
	}
	if g.SideboardingOpen() || view.Match.Game != 2 || view.Match.Sideboarding != nil || !view.MulligansOpen {
		t.Fatalf("after both readied up: match %+v", view.Match)
	}
	if view.StartingSeat != 0 {
		t.Errorf("starting seat %d; Bob chose to draw, so Alice plays first", view.StartingSeat)
	}
	if n := undoDepth(room); n != 0 {
		t.Errorf("%d undo entries reach back into sideboarding", n)
	}

	// Game 2: an ordinary action is undoable again, inside the game.
	if _, err := send(room, alice, actions.TypeKeepHand, nil); err != nil {
		t.Fatalf("keep_hand in game 2: %v", err)
	}
	if undoDepth(room) != 1 {
		t.Errorf("undo depth in game 2 = %d, want 1", undoDepth(room))
	}

	// Conceding the match ends the table whatever the score.
	view, err = send(room, alice, actions.TypeConcedeMatch, nil)
	if err != nil {
		t.Fatalf("concede_match: %v", err)
	}
	if view.State != "ended" || !view.Match.Over || view.Match.WinnerSeat == nil || *view.Match.WinnerSeat != 1 ||
		view.Match.ConcededSeat == nil || *view.Match.ConcededSeat != 0 {
		t.Fatalf("after concede_match: state %q, match %+v", view.State, view.Match)
	}
}

// A Commander table at the room: no gate, no match, undo as it was.
func TestCommanderRoomIsUntouchedByMatches(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := 0; j < 30; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	a, b := g.Seats[0].ID, g.Seats[1].ID
	for _, p := range []uuid.UUID{a, b} {
		if _, err := send(room, p, actions.TypeKeepHand, nil); err != nil {
			t.Fatal(err)
		}
	}
	depth := undoDepth(room)
	if depth == 0 {
		t.Fatal("no undo entries at a Commander table")
	}
	if _, err := send(room, a, actions.TypeConcedeMatch, nil); !errors.Is(err, game.ErrNoMatch) {
		t.Errorf("concede_match at a Commander table: %v, want ErrNoMatch", err)
	}
	view, err := send(room, b, actions.TypeConcede, nil)
	if err != nil {
		t.Fatal(err)
	}
	if view.State != "ended" || view.Match != nil || view.Outcome == nil {
		t.Errorf("state %q, match %+v, outcome %+v", view.State, view.Match, view.Outcome)
	}
	if got := undoDepth(room); got != depth+1 {
		t.Errorf("undo depth %d after the game ended, want %d: a Commander table keeps its history", got, depth+1)
	}
}
