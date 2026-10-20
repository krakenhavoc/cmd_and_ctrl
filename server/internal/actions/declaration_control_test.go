package actions

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaration_control_test.go — #4154: routing under a "you choose which
// creatures attack / block" effect (Master Warcraft). Alice is the
// active player, Bob chooses the declarations, Carol defends.

func newDeclarationControlGame(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID) {
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
		ok = g.BeginDeclarationControlForEffect(game.DeclarationControl{
			Declarer: bob, Source: uuid.New(), SourceName: "Master Warcraft", Attackers: true, Blockers: true,
		})
	})
	if !ok {
		t.Fatal("BeginDeclarationControlForEffect refused")
	}
	return g, alice, bob, carol
}

func seatOf(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// Bob makes Alice's attack declaration as Alice, and nothing else; Alice
// makes none of it, then decides her own half.
func TestDeclarationControlRoutesTheAttackDeclaration(t *testing.T) {
	g, alice, bob, carol := newDeclarationControlGame(t)
	bear := pushCreature(t, g, seatOf(g, alice))
	advanceTo(t, g, game.StepDeclareAttackers)

	declare := func(caller, player uuid.UUID, target uuid.UUID) error {
		return Dispatch(g, Action{Type: TypeDeclareAttacker, Player: player, Caller: caller,
			Params: params(t, map[string]string{"attacker": bear, "target": target.String()})})
	}
	if err := declare(alice, alice, carol); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("Alice declaring in Bob's half: %v, want ErrPlayerControlled", err)
	}
	if err := Dispatch(g, Action{Type: TypeEndTurn, Player: alice, Caller: bob}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("Bob ending Alice's turn: %v, want ErrPlayerControlled", err)
	}
	if err := declare(bob, alice, carol); err != nil {
		t.Fatalf("Bob declaring Alice's attacker: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: alice, Caller: bob}); err != nil {
		t.Fatalf("Bob finishing his half: %v", err)
	}
	if h := g.AttackHandoff; h == nil || h.Stage != game.AttackHandoffTargets {
		t.Fatalf("handoff = %+v, want the targets half", h)
	}
	// Alice's half is hers.
	if err := declare(bob, alice, bob); err == nil {
		t.Fatal("Bob re-pointed Alice's attacker in her half")
	}
	if err := declare(alice, alice, bob); err != nil {
		t.Fatalf("Alice re-pointing her attacker: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: alice, Caller: alice}); err != nil {
		t.Fatalf("Alice's accepting pass: %v", err)
	}
	if h := g.AttackHandoff; h == nil || h.Stage != game.AttackHandoffChosen {
		t.Fatalf("handoff = %+v, want chosen", h)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == bear && c.AttackingTarget != bob {
			t.Errorf("bear attacks %v, want Bob (Alice's choice of target)", c.AttackingTarget)
		}
	}
}

// Bob makes Carol's block declaration as Carol; Carol makes none of it.
func TestDeclarationControlRoutesTheBlockDeclaration(t *testing.T) {
	g, alice, bob, carol := newDeclarationControlGame(t)
	bear := pushCreature(t, g, seatOf(g, alice))
	wall := pushCreature(t, g, seatOf(g, carol))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := Dispatch(g, Action{Type: TypeDeclareAttacker, Player: alice, Caller: bob,
		Params: params(t, map[string]string{"attacker": bear, "target": carol.String()})}); err != nil {
		t.Fatal(err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: alice, Caller: bob}); err != nil {
		t.Fatal(err)
	}
	if err := Dispatch(g, Action{Type: TypePassPriority, Player: alice, Caller: alice}); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)

	block := func(caller uuid.UUID) error {
		return Dispatch(g, Action{Type: TypeDeclareBlocker, Player: carol, Caller: caller,
			Params: params(t, map[string]string{"blocker": wall, "attacker": bear})})
	}
	if err := block(carol); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("Carol blocking for herself: %v, want ErrPlayerControlled", err)
	}
	if err := Dispatch(g, Action{Type: TypeFinishBlocks, Player: carol, Caller: carol}); !errors.Is(err, game.ErrPlayerControlled) {
		t.Fatalf("Carol finishing her own declaration: %v, want ErrPlayerControlled", err)
	}
	if err := block(bob); err != nil {
		t.Fatalf("Bob blocking for Carol: %v", err)
	}
	if err := Dispatch(g, Action{Type: TypeFinishBlocks, Player: carol, Caller: bob}); err != nil {
		t.Fatalf("Bob finishing Carol's declaration: %v", err)
	}
	if g.BlockDeclarationStatusOf(carol) != game.BlockDeclarationDeclared {
		t.Fatal("Carol's declaration is still pending")
	}
	if g.ControllerOfPlayer(carol) != carol {
		t.Error("Carol is still decided for after her declaration")
	}
}
