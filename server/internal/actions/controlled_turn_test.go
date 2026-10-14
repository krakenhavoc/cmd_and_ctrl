package actions

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_turn_test.go — ADR 0167 §4: routing while one player plays
// another's whole turn (CR 723.1, Emrakul, the Promised End). Alice
// (seat 0) controls Bob (seat 1) during Bob's turn; Carol (seat 2) is
// neither.

// newTurnControlGame is a three-seat game on Bob's controlled turn, at
// his upkeep with priority.
func newTurnControlGame(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID) {
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
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	// Seat order from the active seat: Alice, Bob, Carol.
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginTurnControlForEffect(game.TurnControl{
			Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Emrakul, the Promised End",
		})
	})
	if !ok {
		t.Fatal("BeginTurnControlForEffect refused")
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Seats[g.Turn.ActiveSeat].ID != bob || g.ControllerOfPlayer(bob) != alice {
		t.Fatal("setup: not on Bob's controlled turn")
	}
	return g, alice, bob, carol
}

func holder(g *game.Game) uuid.UUID {
	if g.Turn.PriorityHolder < 0 {
		return uuid.Nil
	}
	return g.Seats[g.Turn.PriorityHolder].ID
}

// The rules verbs pass, sent by the controller as the controlled player,
// and run as that player.
func TestTurnControlRulesVerbsRunAsTheControlledPlayer(t *testing.T) {
	g, alice, bob, _ := newTurnControlGame(t)
	if holder(g) != bob {
		t.Fatalf("setup: priority with %v, want Bob", holder(g))
	}
	// A pass sent as Alice herself is hers, and she does not hold
	// priority.
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: alice, Caller: alice}); !errors.Is(err, ErrNotPriorityHolder) {
		t.Fatalf("Alice passing as herself: %v, want ErrNotPriorityHolder", err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: alice}); err != nil {
		t.Fatalf("Alice passing for Bob: %v", err)
	}
	if holder(g) == bob {
		t.Fatal("Bob still holds priority after Alice passed for him")
	}

	// A land played from Bob's hand by Alice, through cast_spell: it is
	// Bob's land (CR 723.3).
	advanceTo(t, g, game.StepPrecombatMain)
	land := game.NewCard("Forest", bob)
	land.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() { g.PlayerByIDForEffect(bob).Hand.PushTop(land) })
	a := Action{Type: TypeCastSpell, Player: bob, Caller: alice,
		Params: params(t, map[string]any{"instance_id": land.InstanceID.String()})}
	if err := Dispatch(g, a); err != nil {
		t.Fatalf("Alice playing Bob's land: %v", err)
	}
	c, ok := g.LookupCardForEffect(land.InstanceID)
	if !ok || !g.Battlefield.Contains(land.InstanceID) || c.Controller != bob {
		t.Fatalf("the land is not on the battlefield under Bob's control: %+v", c)
	}

	// end_turn is the dock's turn walk, for Bob as the active player
	// (requireActivePlayer reads the controlled seat).
	if err := Dispatch(g, Action{Type: TypeEndTurn, Player: bob, Caller: alice}); err != nil {
		t.Fatalf("Alice ending Bob's turn: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypeEndTurn, Player: alice, Caller: alice}); !errors.Is(err, ErrNotActivePlayer) {
		t.Errorf("Alice ending the turn as herself: %v, want ErrNotActivePlayer", err)
	}
}

// The controlled seat decides nothing but to concede (CR 723.5, 723.6),
// and the controller can't concede for them.
func TestTurnControlControlledSeatIsRefusedButMayConcede(t *testing.T) {
	g, alice, bob, _ := newTurnControlGame(t)
	for _, a := range []Action{
		{Type: TypePassPriority, Player: bob, Caller: bob},
		{Type: TypeEndTurn, Player: bob, Caller: bob},
		{Type: TypeDrawCard, Player: bob, Caller: bob},
		{Type: TypeSetTriggerOrderPreference, Player: bob, Caller: bob, Params: []byte(`{"trigger_order":"always"}`)},
	} {
		if err := Dispatch(g, a); !errors.Is(err, game.ErrPlayerControlled) {
			t.Errorf("Bob's own %s: %v, want ErrPlayerControlled", a.Type, err)
		}
	}
	if err := Dispatch(g, Action{Type: TypeConcede, Player: bob, Caller: alice}); !errors.Is(err, ErrPlayerCallerMismatch) {
		t.Errorf("Alice conceding for Bob: %v, want ErrPlayerCallerMismatch", err)
	}
	if err := Dispatch(g, Action{Type: TypeConcede, Player: bob, Caller: bob}); err != nil {
		t.Fatalf("Bob conceding: %v", err)
	}
	if p := g.PlayerByID(bob); p == nil || !p.Eliminated {
		t.Error("Bob is still in the game after conceding")
	}
}

// The sandbox verbs are closed to both players of the turn (CR 723.5b),
// as either player. A third player and the admin keep them.
func TestTurnControlSandboxVerbsAreClosedToBothPlayers(t *testing.T) {
	g, alice, bob, carol := newTurnControlGame(t)
	for _, a := range []Action{
		{Type: TypeDrawCard, Player: bob, Caller: alice},
		{Type: TypeDrawCard, Player: alice, Caller: alice},
		{Type: TypeChangeLife, Player: bob, Caller: alice, Params: []byte(`{"delta":-1}`)},
		{Type: TypePassTurn, Player: bob, Caller: alice},
		{Type: TypeUntapAll, Player: alice, Caller: alice},
	} {
		if err := Dispatch(g, a); !errors.Is(err, game.ErrPlayerControlled) {
			t.Errorf("%s from Alice for %v: %v, want ErrPlayerControlled", a.Type, a.Player, err)
		}
	}
	// A setting is not a sandbox verb: Alice keeps her own.
	if err := Dispatch(g, Action{Type: TypeSetTriggerOrderPreference, Player: alice, Caller: alice,
		Params: []byte(`{"trigger_order":"always"}`)}); err != nil {
		t.Errorf("Alice's own setting: %v", err)
	}
	hand := g.PlayerByID(carol).Hand.Size()
	if err := Dispatch(g, Action{Type: TypeDrawCard, Player: carol, Caller: carol}); err != nil {
		t.Errorf("Carol's draw_card: %v", err)
	}
	if g.PlayerByID(carol).Hand.Size() != hand+1 {
		t.Error("Carol did not draw")
	}
	bobHand := g.PlayerByID(bob).Hand.Size()
	if err := Dispatch(g, Action{Type: TypeDrawCard, Player: bob, Caller: uuid.Nil, Admin: true}); err != nil {
		t.Errorf("the admin's draw_card for Bob: %v", err)
	}
	if g.PlayerByID(bob).Hand.Size() != bobHand+1 {
		t.Error("the admin's draw for Bob did not happen")
	}
}

// A third player can't act for the controlled player, and the
// controller can't act for a seat they don't control.
func TestTurnControlOnlyTheControllerActsForThem(t *testing.T) {
	g, _, bob, carol := newTurnControlGame(t)
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: carol}); err == nil {
		t.Fatal("Carol passed for Bob")
	}
	if holder(g) != bob {
		t.Error("priority moved")
	}
}
