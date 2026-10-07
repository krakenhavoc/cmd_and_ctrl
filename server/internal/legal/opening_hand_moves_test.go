package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// opening_hand_moves_test.go — ADR 0133. The questions an opening-hand
// action asks are the chained-choice kinds (confirm, then choose_cards),
// so the seat owing one is offered its answers by the ordinary choice
// enumerator and nobody else is offered anything — not even a pass, the
// table being parked (#499, #544).

const (
	oracleLeylineOfSanctity = "492e0e6c-8c27-4376-938b-f8a8b6205810"
	oracleGemstoneCaverns   = "c0adbddc-b070-4c5f-afe0-0474c72a9251"
)

// openingHandTable starts a two-seat game with the mulligan window
// open and gives seat 1 the named cards as its whole opening hand.
func openingHandTable(t *testing.T, hand ...game.Card) *game.Game {
	t.Helper()
	g := newTableMulligans(t)
	p := g.Seats[1]
	clearHand(p)
	for _, c := range hand {
		handCard(p, c)
	}
	return g
}

func keepBoth(t *testing.T, g *game.Game) {
	t.Helper()
	for i := range g.Seats {
		if err := g.KeepHand(g.Seats[(g.StartingSeat+i)%len(g.Seats)].ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
}

func TestOpeningHandOfferIsAnsweredByTheSeatThatOwesIt(t *testing.T) {
	g := openingHandTable(t, game.Card{Name: "Leyline of Sanctity", TypeLine: "Enchantment", ManaCost: "{2}{W}{W}", OracleID: oracleLeylineOfSanctity})
	keepBoth(t, g)

	owed, other := g.Seats[1], g.Seats[0]
	if got := legal.EnumerateFor(g, other.ID); len(got) != 0 {
		t.Fatalf("a seat that owes nothing is offered %v while the table is parked", labels(got))
	}
	moves := legal.EnumerateFor(g, owed.ID)
	if countKind(moves, legal.KindChoice) != 2 {
		t.Fatalf("want accept and decline, got %v", labels(moves))
	}
	if !hasLabel(moves, "Begin the game with Leyline of Sanctity on the battlefield?: Begin the game with it") ||
		!hasLabel(moves, "Begin the game with Leyline of Sanctity on the battlefield?: Keep it in my hand") {
		t.Errorf("the answers do not read like the card: %v", labels(moves))
	}
	var alwaysLegal int
	for _, m := range moves {
		if m.AlwaysLegal {
			alwaysLegal++
		}
	}
	if alwaysLegal != 1 {
		t.Errorf("%d moves marked always-legal, want exactly the decline", alwaysLegal)
	}
	dispatchAll(t, g, owed.ID, moves)
}

// Gemstone Caverns asks twice: yes or no, then which card to exile. The
// second question is offered one answer per card left in hand.
func TestGemstoneCavernsExilePickIsEnumerated(t *testing.T) {
	g := openingHandTable(t,
		game.Card{Name: "Gemstone Caverns", TypeLine: "Legendary Land", OracleID: oracleGemstoneCaverns},
		game.Card{Name: "Spare One", TypeLine: "Sorcery", ManaCost: "{1}{R}"},
		game.Card{Name: "Spare Two", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
	)
	// Seat 1 is not the starting player at a two-seat table started
	// with seat 0 first, so the Caverns is offered to it.
	keepBoth(t, g)
	owed := g.Seats[1]
	moves := legal.EnumerateFor(g, owed.ID)
	var accept *legal.Move
	for i := range moves {
		if moves[i].Kind == legal.KindChoice && !moves[i].AlwaysLegal {
			accept = &moves[i]
		}
	}
	if accept == nil {
		t.Fatalf("no accept move in %v", labels(moves))
	}
	dispatchOne(t, g, owed.ID, *accept)

	picks := legal.EnumerateFor(g, owed.ID)
	if countKind(picks, legal.KindChoice) < 2 {
		t.Fatalf("the exile pick offers %v, want one answer per card left in hand", labels(picks))
	}
	dispatchAll(t, g, owed.ID, picks)
	if got := legal.EnumerateFor(g, g.Seats[0].ID); len(got) != 0 {
		t.Errorf("the other seat is offered %v during the exile pick", labels(got))
	}
}
