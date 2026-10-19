package protocol

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// declaration_control_view_test.go — #4154: what the view and the
// declarer's legal moves carry under a "you choose which creatures
// attack / block" effect (Master Warcraft). Alice is active, Bob
// chooses, Carol defends.

func declarationControlTable(t *testing.T) (g *game.Game, alice, bob, carol uuid.UUID, bear uuid.UUID) {
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
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob, carol = g.Seats[a].ID, g.Seats[(a+1)%3].ID, g.Seats[(a+2)%3].ID
	c := game.NewCard("Bear", alice)
	c.TypeLine = "Creature — Bear"
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	bear = c.InstanceID
	g.WithWriteLock(func() {
		g.BeginDeclarationControlForEffect(game.DeclarationControl{
			Declarer: bob, Source: uuid.New(), SourceName: "Master Warcraft", Attackers: true, Blockers: true,
		})
	})
	return g, alice, bob, carol, bear
}

// The effect is public all turn. While Bob chooses Alice's attackers he
// is listed as making her declaration, offered her attack moves and the
// finishing pass as her, and shown nothing of her hand; Alice is offered
// nothing. In her half the handoff says so and the moves are hers.
func TestDeclarationControlViewAndTheDeclarersMoves(t *testing.T) {
	g, alice, bob, carol, bear := declarationControlTable(t)
	for _, viewer := range []string{alice.String(), bob.String(), carol.String(), SpectatorViewerID} {
		v := ViewOfGameFor(g, viewer)
		if len(v.DeclarationControls) != 1 || v.DeclarationControls[0].Declarer != bob.String() ||
			!v.DeclarationControls[0].Attackers || !v.DeclarationControls[0].Blockers || v.DeclarationControls[0].Scope != "turn" {
			t.Fatalf("viewer %s: declaration_controls = %+v", viewer, v.DeclarationControls)
		}
		if len(v.PlayerControls) != 0 {
			t.Fatalf("viewer %s: player_controls before combat = %+v", viewer, v.PlayerControls)
		}
	}
	stepToSeatStep(t, g, alice, game.StepDeclareAttackers)

	full := ViewOfGame(g)
	bv := FilterViewFor(full, bob.String())
	if len(bv.PlayerControls) != 1 || bv.PlayerControls[0].Until != "declarations" ||
		bv.PlayerControls[0].Controller != bob.String() || bv.PlayerControls[0].Controlled != alice.String() {
		t.Fatalf("player_controls = %+v, want Bob making Alice's declaration", bv.PlayerControls)
	}
	// Choosing her attackers is not controlling her: her hand stays hers.
	for _, c := range seatViewOf(t, bv, alice).Hand.Cards {
		if c.KnownByYou {
			t.Fatalf("Bob knows %s in Alice's hand", c.Name)
		}
	}
	var attack, pass bool
	for _, m := range bv.LegalMoves {
		if m.Player != alice {
			continue
		}
		switch m.Type {
		case legal.TypeDeclareAttacker:
			attack = attack || m.Source == bear
		case legal.TypePassPriority:
			pass = true
		default:
			t.Errorf("Bob is offered %s as Alice", m.Type)
		}
	}
	if !attack || !pass {
		t.Fatalf("Bob's moves as Alice: attack %v pass %v (%+v)", attack, pass, bv.LegalMoves)
	}
	if d := bv.ControlledLegalActions[alice.String()]; d == nil || !d.Pass || d.Sources[bear.String()] == nil {
		t.Errorf("controlled_legal_actions[alice] = %+v, want the bear's attack and the pass", d)
	}
	if av := FilterViewFor(full, alice.String()); len(av.LegalMoves) != 0 {
		t.Errorf("Alice is offered %d moves in Bob's half", len(av.LegalMoves))
	}

	if err := g.DeclareAttacker(bear, carol); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	full = ViewOfGame(g)
	for _, viewer := range []string{alice.String(), carol.String()} {
		v := FilterViewFor(full, viewer)
		if v.AttackHandoff == nil || v.AttackHandoff.Stage != "targets" || v.AttackHandoff.Declarer != bob.String() {
			t.Fatalf("viewer %s: attack_handoff = %+v", viewer, v.AttackHandoff)
		}
		if len(v.PlayerControls) != 0 {
			t.Fatalf("viewer %s: player_controls in Alice's half = %+v", viewer, v.PlayerControls)
		}
	}
	av := FilterViewFor(full, alice.String())
	var repoint, passOK bool
	for _, m := range av.LegalMoves {
		switch m.Type {
		case legal.TypeDeclareAttacker:
			repoint = repoint || m.Source == bear
		case legal.TypePassPriority:
			passOK = true
		}
	}
	if !repoint || !passOK {
		t.Fatalf("Alice's half: re-point %v pass %v (%+v)", repoint, passOK, av.LegalMoves)
	}
}
