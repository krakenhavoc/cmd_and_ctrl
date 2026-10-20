package actions

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// controlled_combat_test.go — #3999: routing while one player plays
// another's combat phase (Secret of Bloodbending, CR 723.2). Alice
// controls Bob during Bob's next combat phase; Carol is neither.

// newCombatControlGame is a three-seat game at the start of Bob's turn,
// with Alice's combat control over Bob waiting and a creature of Bob's
// on the battlefield.
func newCombatControlGame(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID, attacker string) {
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
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginCombatControlForEffect(game.CombatControl{
			Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Secret of Bloodbending",
		})
	})
	if !ok {
		t.Fatal("BeginCombatControlForEffect refused")
	}
	attacker = pushCreature(t, g, g.Seats[(a+1)%3])
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Seats[g.Turn.ActiveSeat].ID != bob {
		t.Fatal("setup: not on Bob's turn")
	}
	return g, alice, bob, carol, attacker
}

// Before the combat phase Bob decides for himself; during it Alice
// passes and attacks for him, and Bob and Carol cannot; after it the
// control is over.
func TestCombatControlRoutesOnlyDuringTheCombatPhase(t *testing.T) {
	g, alice, bob, carol, attacker := newCombatControlGame(t)
	advanceTo(t, g, game.StepPrecombatMain)
	if g.ControllerOfPlayer(bob) != bob {
		t.Fatal("Bob is controlled before his combat phase")
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: alice}); err == nil {
		t.Fatal("Alice passed for Bob before his combat phase")
	}

	advanceTo(t, g, game.StepDeclareAttackers)
	if g.ControllerOfPlayer(bob) != alice {
		t.Fatal("setup: Bob's combat phase is not Alice's")
	}
	// Alice attacks with Bob's creature, at herself (CR 723.5's
	// example): it is still Bob's creature.
	if err := Dispatch(g, Action{Type: TypeDeclareAttacker, Player: bob, Caller: alice,
		Params: params(t, map[string]string{"attacker": attacker, "target": alice.String()})}); err != nil {
		t.Fatalf("Alice declaring Bob's attacker: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == attacker && (c.AttackingTarget != alice || c.Controller != bob) {
			t.Errorf("attacker = target %v controller %v, want attacking Alice under Bob", c.AttackingTarget, c.Controller)
		}
	}
	// Bob decides nothing of this phase (CR 723.5), and Carol is not
	// his controller.
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: bob}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Errorf("Bob passing for himself: %v, want ErrPlayerControlled", err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: carol}); err == nil {
		t.Error("Carol passed for Bob")
	}
	// end_turn would walk Bob's turn past the phase: refused.
	if err := Dispatch(g, Action{Type: TypeEndTurn, Player: bob, Caller: alice}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Errorf("Alice ending Bob's turn: %v, want ErrPlayerControlled", err)
	}
	// The sandbox is closed to Alice, as Bob and as herself (CR 723.5b).
	for _, a := range []Action{
		{Type: TypeDrawCard, Player: bob, Caller: alice},
		{Type: TypeDrawCard, Player: alice, Caller: alice},
		{Type: TypeClearCombat, Player: bob, Caller: alice},
	} {
		if err := Dispatch(g, a); !errors.Is(err, game.ErrPlayerControlled) {
			t.Errorf("%s from Alice for %v: %v, want ErrPlayerControlled", a.Type, a.Player, err)
		}
	}
	if h := holder(g); h == bob {
		if err := Dispatch(g, Action{Type: TypePassPriority, Player: bob, Caller: alice}); err != nil {
			t.Fatalf("Alice passing for Bob: %v", err)
		}
	}

	advanceTo(t, g, game.StepPostcombatMain)
	if g.ControllerOfPlayer(bob) != bob {
		t.Fatal("the control outlived Bob's combat phase")
	}
	if err := Dispatch(g, Action{Type: TypeDrawCard, Player: alice, Caller: alice}); err != nil {
		t.Errorf("Alice's own sandbox after the combat phase: %v", err)
	}
}
