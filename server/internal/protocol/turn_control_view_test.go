package protocol

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// turn_control_view_test.go — ADR 0167 §7 and §8: what the view and the
// controller's legal moves carry while one player plays another's turn.

func turnControlTable(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID) {
	t.Helper()
	g = game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	g.WithWriteLock(func() {
		g.BeginTurnControlForEffect(game.TurnControl{
			Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Emrakul, the Promised End",
		})
	})
	return g, alice, bob, carol
}

func TestScheduledPlayerControlIsPublic(t *testing.T) {
	g, alice, bob, carol := turnControlTable(t)
	for _, viewer := range []string{alice.String(), bob.String(), carol.String(), SpectatorViewerID} {
		v := ViewOfGameFor(g, viewer)
		if len(v.ScheduledPlayerControls) != 1 {
			t.Fatalf("viewer %s: scheduled_player_controls = %+v", viewer, v.ScheduledPlayerControls)
		}
		s := v.ScheduledPlayerControls[0]
		if s.Controller != alice.String() || s.Controlled != bob.String() || s.SourceName != "Emrakul, the Promised End" {
			t.Errorf("viewer %s: %+v", viewer, s)
		}
		if len(v.PlayerControls) != 0 {
			t.Errorf("viewer %s: a waiting control is listed as live: %+v", viewer, v.PlayerControls)
		}
	}
	// It grants no visibility: Alice does not see Bob's hand yet.
	v := ViewOfGameFor(g, alice.String())
	for _, c := range seatViewOf(t, v, bob).Hand.Cards {
		if c.KnownByYou {
			t.Fatalf("Alice knows %s in Bob's hand before the turn", c.Name)
		}
	}
}

func TestTurnControlViewAndTheControllersMoves(t *testing.T) {
	g, alice, bob, carol := turnControlTable(t)
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	full := ViewOfGame(g)
	av := FilterViewFor(full, alice.String())
	if len(av.PlayerControls) != 1 {
		t.Fatalf("player_controls = %+v", av.PlayerControls)
	}
	pc := av.PlayerControls[0]
	if pc.Controller != alice.String() || pc.Controlled != bob.String() || pc.Until != "turn" || pc.SourceName != "Emrakul, the Promised End" {
		t.Errorf("player_controls[0] = %+v", pc)
	}
	if len(av.ScheduledPlayerControls) != 0 {
		t.Errorf("a begun control is still scheduled: %+v", av.ScheduledPlayerControls)
	}
	// CR 723.4: Alice sees Bob's hand.
	for _, c := range seatViewOf(t, av, bob).Hand.Cards {
		if !c.KnownByYou {
			t.Errorf("Alice does not know %s in Bob's hand", c.Name)
		}
	}
	// Bob holds priority. Alice's list is his moves, Player set to him;
	// her own digest is empty and his is under his ID.
	var forBob int
	for _, m := range av.LegalMoves {
		if m.Player != bob {
			t.Errorf("Alice's list has a move for %v while Bob holds priority: %s", m.Player, m.Label)
		}
		forBob++
	}
	if forBob == 0 {
		t.Fatal("Alice is offered nothing for Bob")
	}
	if av.LegalActions != nil && av.LegalActions.Pass {
		t.Error("Alice's own legal_actions says she may pass")
	}
	if d := av.ControlledLegalActions[bob.String()]; d == nil || !d.Pass {
		t.Errorf("controlled_legal_actions[bob] = %+v, want a pass", d)
	}
	// Bob is offered nothing, and neither is Carol.
	for _, id := range []uuid.UUID{bob, carol} {
		if v := FilterViewFor(full, id.String()); len(v.LegalMoves) != 0 || v.ControlledLegalActions != nil {
			t.Errorf("%v: legal_moves %d, controlled_legal_actions %v", id, len(v.LegalMoves), v.ControlledLegalActions)
		}
	}
	// The enumerator agrees with the wire.
	if got := legal.EnumerateFor(g, bob); len(got) != 0 {
		t.Errorf("Bob's enumeration = %d moves", len(got))
	}
}

// The wire cap is per player: Alice's own moves can't fold Bob's away.
func TestCapLegalMovesIsPerPlayer(t *testing.T) {
	alice, bob := uuid.New(), uuid.New()
	src := uuid.New()
	var moves []LegalMoveView
	for i := 0; i < legalMovesWireCap+10; i++ {
		moves = append(moves, LegalMoveView{Type: "cast_spell", Player: alice, Kind: legal.KindCast, Source: src, Label: fmt.Sprintf("a%d", i)})
	}
	for i := 0; i < 5; i++ {
		moves = append(moves, LegalMoveView{Type: "cast_spell", Player: bob, Kind: legal.KindCast, Source: src, Label: fmt.Sprintf("b%d", i)})
	}
	out, cut := capLegalMoves(moves)
	if !cut {
		t.Fatal("not cut")
	}
	var a, b int
	for _, m := range out {
		switch m.Player {
		case alice:
			a++
		case bob:
			b++
		}
	}
	// Bob's five fit under the cap on their own and are kept whole;
	// Alice's are folded to one per source and kind.
	if a != 1 || b != 5 {
		t.Errorf("after the cap: %d of Alice's, %d of Bob's; want 1 and 5", a, b)
	}
	if !strings.HasPrefix(out[0].Label, "a") {
		t.Errorf("order changed: first is %q", out[0].Label)
	}
}
