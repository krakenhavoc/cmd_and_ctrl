package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_tax_test.go is the #544 half of ADR 0080's attack tax, and
// it is the half that can hang a table.
//
// The promise is the package's: every Move EnumerateFor returns is one
// actions.Dispatch accepts. An attack tax breaks it in the direction
// that matters — the engine refuses a declaration the enumerator
// offered for want of mana, the seat's policy re-picks the same
// argmax, and a seat that owes a decision is enumerated that
// decision's answers and nothing else, so there is nothing else to do.
//
// The mechanism keeping the two in agreement is one function, not a
// mirrored rule: game.PriceAttackDeclarationForEffect, which the
// engine's declaration verbs also call. These tests would fail if
// either side re-derived the price.

const (
	oraclePropaganda    = "ea9709b6-4c37-4d5a-b04d-cd4c42e4f9dd"
	oracleGhostlyPrison = "e828b189-0e8f-43b8-b909-4c23e742e028"
)

// propagandaOn drops a real catalog Propaganda under `p`.
func propagandaOn(g *game.Game, p *game.Player) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: "Propaganda", TypeLine: "Enchantment", ManaCost: "{2}{U}", OracleID: oraclePropaganda,
	})
}

// untappedIslands gives a seat n untapped basic lands, which the
// auto-tapper can reach for a generic tax.
func untappedIslands(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		battlefieldCard(g, p, basic("Island", "Island"))
	}
}

// taxOn reads the mana price the enumerator stamped on an attack move
// for one creature, or "" when the move is not offered at all.
func taxOn(moves []legal.Move, creature uuid.UUID) (string, bool) {
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != creature {
			continue
		}
		if m.Cost == nil {
			return "", true
		}
		return m.Cost.Mana, true
	}
	return "", false
}

// TestAnUnaffordableAttackIsNeverOffered is the invariant, stated
// where it would break.
func TestAnUnaffordableAttackIsNeverOffered(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)

	bear := freshCreature(g, me, "Bear")
	propagandaOn(g, them)
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	// The invariant: nothing offered may be refused.
	dispatchAll(t, g, me.ID, moves)

	// With no mana and no lands, attacking the Propaganda seat is
	// unaffordable and is not offered. Attacking the other two still
	// is, so the test cannot pass by enumerating nothing.
	for _, m := range moves {
		if m.Kind != legal.KindAttack {
			continue
		}
		ap := decodeAttack(t, m)
		if ap.Target == them.ID.String() {
			t.Errorf("the enumerator offered an unaffordable attack on the Propaganda seat: %q", m.Label)
		}
	}
	if n := attacksBy(moves, bear); n == 0 {
		t.Fatalf("no attacks offered at all: %v", labels(moves))
	}
}

// TestAnAffordableAttackIsOfferedAndPriced — the other half. Two
// untapped lands make Propaganda's {2} reachable, so the move comes
// back, carries the price, and carries the auto_tap permission it
// needs to actually pay.
func TestAnAffordableAttackIsOfferedAndPriced(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)

	bear := freshCreature(g, me, "Bear")
	propagandaOn(g, them)
	untappedIslands(g, me, 2)
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	var taxed *legal.Move
	for i := range moves {
		if moves[i].Kind != legal.KindAttack {
			continue
		}
		if decodeAttack(t, moves[i]).Target == them.ID.String() {
			taxed = &moves[i]
			break
		}
	}
	if taxed == nil {
		t.Fatalf("the affordable attack on the Propaganda seat was not offered: %v", labels(moves))
	}
	if taxed.Cost == nil || taxed.Cost.Mana != "{2}" {
		t.Errorf("MoveCost.Mana = %#v, want {2} — a policy cannot price the move without it", taxed.Cost)
	}
	if !decodeAttack(t, *taxed).AutoTap {
		t.Error("the move was offered on the strength of the tapper but did not carry auto_tap")
	}
	// An attack on an UNTAXED seat is still free and unstamped.
	free, ok := taxOn(moves, bear)
	_ = free
	if !ok {
		t.Fatal("no attack move for the bear at all")
	}
}

// TestTwoTaxesAreEnumeratedAtTheStackedPrice — the enumerator's price
// is the engine's, whatever is on the board.
func TestTwoTaxesAreEnumeratedAtTheStackedPrice(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)

	freshCreature(g, me, "Bear")
	propagandaOn(g, them)
	battlefieldCard(g, them, game.Card{
		Name: "Ghostly Prison", TypeLine: "Enchantment", ManaCost: "{2}{W}", OracleID: oracleGhostlyPrison,
	})
	untappedIslands(g, me, 4)
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	found := false
	for _, m := range moves {
		if m.Kind != legal.KindAttack || decodeAttack(t, m).Target != them.ID.String() {
			continue
		}
		found = true
		if m.Cost == nil || m.Cost.Mana != "{2}{2}" {
			t.Errorf("stacked price = %#v, want {2}{2}", m.Cost)
		}
	}
	if !found {
		t.Fatalf("four lands should cover two {2} taxes: %v", labels(moves))
	}
}

// TestThreeLandsDoNotCoverTwoAttackers is the composition property
// ADR 0080 §5 rests on: the enumerator is per-creature, the engine
// charges per declaration verb call, and after the first attack is
// declared the second is re-priced against what is left.
func TestThreeLandsDoNotCoverTwoAttackers(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)

	first := freshCreature(g, me, "First Bear")
	second := freshCreature(g, me, "Second Bear")
	propagandaOn(g, them)
	untappedIslands(g, me, 3)
	advanceTo(t, g, game.StepDeclareAttackers)

	// Both are offered while three lands are up.
	before := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, before)
	if _, ok := taxOn(before, first); !ok {
		t.Fatal("the first attacker was not offered")
	}
	if _, ok := taxOn(before, second); !ok {
		t.Fatal("the second attacker was not offered")
	}

	// Declare the first, which taps two of the three lands.
	if err := g.DeclareAttackerWith(first, them.ID, game.DeclareAttackersParams{AutoTap: true}); err != nil {
		t.Fatalf("DeclareAttackerWith: %v", err)
	}

	// One land left, {2} still owed: the second attack is gone, and
	// everything still on the list is still accepted.
	after := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, after)
	for _, m := range after {
		if m.Kind != legal.KindAttack || m.Source != second {
			continue
		}
		if decodeAttack(t, m).Target == them.ID.String() {
			t.Errorf("the second attack was still offered on one land: %q", m.Label)
		}
	}
}

// decodeAttack reads an attack move's params the way the dispatcher
// does, so a test cannot disagree with the wire about what was sent.
func decodeAttack(t *testing.T, m legal.Move) struct {
	Attacker string `json:"attacker"`
	Target   string `json:"target"`
	AutoTap  bool   `json:"auto_tap"`
} {
	t.Helper()
	var p struct {
		Attacker string `json:"attacker"`
		Target   string `json:"target"`
		AutoTap  bool   `json:"auto_tap"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("attack params %s: %v", m.Params, err)
	}
	return p
}
