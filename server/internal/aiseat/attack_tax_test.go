package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// attack_tax_test.go — the bot half of ADR 0080 (#1063).
//
// An attack tax is the first thing that can make a DECLARATION fail
// for want of mana, and #544's rule is that a bot is never offered a
// move the engine refuses. If the enumerator got that wrong the seat
// would re-pick the same argmax forever, which is the wedge shape the
// whole of S36 was about.
//
// Deliberately NOT behind AISEAT_GAME_TESTS: it plays no whole game
// and runs in milliseconds, and the gate on the whole-game tests is
// exactly why wedges of this shape have reached production before.

const (
	botPropagandaOracle    = "ea9709b6-4c37-4d5a-b04d-cd4c42e4f9dd"
	botGhostlyPrisonOracle = "e828b189-0e8f-43b8-b909-4c23e742e028"
)

// propagandaTable parks a two-seat game on the active seat's declare
// attackers with a 3/3 under the active seat, a real Propaganda under
// the other, and `lands` untapped Mountains for the attacker.
func propagandaTable(t *testing.T, lands int) (*game.Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newSettledTable(t, 31)
	me, them := g.Seats[0], g.Seats[1]

	push := func(owner *game.Player, name, typeLine, oracle string, power, tough int) uuid.UUID {
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: tough,
			Owner: owner.ID, Controller: owner.ID,
		}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	attacker := push(me, "Ogre", "Creature — Ogre", "", 4, 4)
	push(them, "Propaganda", "Enchantment", botPropagandaOracle, 0, 0)
	for i := 0; i < lands; i++ {
		push(me, "Mountain", "Basic Land — Mountain", "", 0, 0)
	}
	advanceToStep(t, g, game.StepDeclareAttackers)
	return g, attacker, them.ID
}

// attackMovesOn counts the attack moves offered against one seat.
func attackMovesOn(moves []legal.Move, target uuid.UUID) int {
	n := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack {
			continue
		}
		var p struct {
			Target string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			continue
		}
		if p.Target == target.String() {
			n++
		}
	}
	return n
}

// TestTheBotIsNotOfferedAnAttackItCannotPayFor.
func TestTheBotIsNotOfferedAnAttackItCannotPayFor(t *testing.T) {
	g, _, them := propagandaTable(t, 0)

	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	if n := attackMovesOn(moves, them); n != 0 {
		t.Fatalf("the bot was offered %d attacks it cannot pay Propaganda's {2} for", n)
	}
	// And the policy does something rather than nothing: with no
	// attack to make it falls through to the general decision, which
	// at worst is the always-legal pass.
	if len(moves) == 0 {
		t.Fatal("the active seat was offered NOTHING at declare attackers — the #544 wedge")
	}
	d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, g.Seats[0].ID.String()),
		Seat:  g.Seats[0].ID,
		Moves: moves,
	})
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("policy picked index %d of %d", d.Index, len(moves))
	}
	if moves[d.Index].Kind == legal.KindAttack {
		t.Errorf("the policy picked an attack anyway: %q", moves[d.Index].Label)
	}
}

// TestTheBotAttacksUnderPropagandaOnceItCanPay — the other half, and
// the one that proves the offer is honest: the policy picks the
// attack, the dispatcher accepts the move byte for byte, and the
// lands the tapper planned are tapped.
func TestTheBotAttacksUnderPropagandaOnceItCanPay(t *testing.T) {
	g, attacker, them := propagandaTable(t, 2)

	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	if n := attackMovesOn(moves, them); n == 0 {
		t.Fatalf("two lands cover Propaganda's {2} but no attack was offered: %v", attackLabels(moves))
	}
	d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, g.Seats[0].ID.String()),
		Seat:  g.Seats[0].ID,
		Moves: moves,
	})
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	pick := moves[d.Index]
	if pick.Kind != legal.KindAttack {
		t.Fatalf("a 4/4 into an empty board for {2} should be worth attacking; the policy took %q", pick.Label)
	}
	// The engine accepts exactly what the enumerator offered.
	if err := actions.Dispatch(g, actions.Action{
		Type:   actions.Type(pick.Type),
		Player: pick.Player,
		Caller: g.Seats[0].ID,
		Params: pick.Params,
	}); err != nil {
		t.Fatalf("the engine refused an enumerated attack: %v", err)
	}
	var tapped int
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == g.Seats[0].ID && c.Name == "Mountain" && c.Tapped {
				tapped++
			}
			if c.InstanceID == attacker && c.AttackingTarget != them {
				t.Errorf("the attacker was not declared against the taxed seat")
			}
		}
	})
	if tapped != 2 {
		t.Errorf("%d Mountains tapped for the tax, want 2", tapped)
	}
}

// TestTheHeuristicPrefersTheUntaxedSeat — the policy weighs the tax
// rather than merely tolerating it. Two identical opponents, one of
// them taxing attacks; the bot has mana for either, and picks the
// free one.
//
// The tax is TWO Propagandas rather than one, and that is the honest
// version of the assertion. A {2} tax is worth 0.6 at the default
// AttackTaxPenalty, which is less than the FocusBonus the aggression
// rotation adds to whichever seat it has settled on — so a single
// Propaganda does not, and should not, override sustained pressure.
// A {4} tax is worth 1.2 and does. Pinning the weaker claim would
// have meant tuning AttackTaxPenalty until the test passed, which is
// how a knob stops meaning anything.
func TestTheHeuristicPrefersTheUntaxedSeat(t *testing.T) {
	g := newRoom(t, 3, 37).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me, taxed, free := g.Seats[0], g.Seats[1], g.Seats[2]

	push := func(owner *game.Player, name, typeLine, oracle string, power, tough int) {
		c := game.Card{
			InstanceID: uuid.New(),
			Name:       name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: tough,
			Owner: owner.ID, Controller: owner.ID,
		}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
	}
	push(me, "Ogre", "Creature — Ogre", "", 4, 4)
	push(taxed, "Propaganda", "Enchantment", botPropagandaOracle, 0, 0)
	push(taxed, "Ghostly Prison", "Enchantment", botGhostlyPrisonOracle, 0, 0)
	for i := 0; i < 4; i++ {
		push(me, "Mountain", "Basic Land — Mountain", "", 0, 0)
	}
	advanceToStep(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	if attackMovesOn(moves, taxed.ID) == 0 || attackMovesOn(moves, free.ID) == 0 {
		t.Fatalf("both seats should be attackable with four lands up: %v", attackLabels(moves))
	}
	d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, me.ID.String()),
		Seat:  me.ID,
		Moves: moves,
	})
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	pick := moves[d.Index]
	if pick.Kind != legal.KindAttack {
		t.Fatalf("the policy did not attack at all: %q", pick.Label)
	}
	var p struct {
		Target string `json:"target"`
	}
	if err := json.Unmarshal(pick.Params, &p); err != nil {
		t.Fatalf("attack params: %v", err)
	}
	if p.Target == taxed.ID.String() {
		t.Errorf("the policy paid Propaganda's {2} when an identical seat was free: %q", pick.Label)
	}
}

// seatIDs is every seat at the table, for AddKnowersAll: a permanent
// is public information.
func seatIDs(g *game.Game) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(g.Seats))
	for _, s := range g.Seats {
		out = append(out, s.ID)
	}
	return out
}

// attackLabels renders a move list for a failure message.
func attackLabels(moves []legal.Move) []string {
	out := make([]string, 0, len(moves))
	for _, m := range moves {
		out = append(out, m.Label)
	}
	return out
}
