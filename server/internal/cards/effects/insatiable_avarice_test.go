package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// insatiable_avarice_test.go — Spree proof card #3 (CR 702.172a, ADR
// 0065's 2026-09-23 amendment): one untargeted bullet beside one
// targeted, colour-priced bullet, proving Spree composes with both
// shapes under one announcement.

const insatiableAvariceOracle = "ad3e705f-da57-4eff-84d7-2072522de988"

// The untargeted tutor bullet alone: a real search choice opens (the
// library is 20 identical matches, so nothing auto-resolves — the
// S22 "you choose" prompt this card shares with Vampiric Tutor), and
// answering it leaves the library the same SIZE — ToTop relocates
// the found card, it does not remove one.
func TestInsatiableAvariceTutorsToTopAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()

	castModal(t, g, "Insatiable Avarice", "Sorcery", insatiableAvariceOracle,
		[]int{0}, nil)
	passPriorityAroundTable(t, g)

	if c := searchChoiceFor(g, me.ID); c == nil {
		t.Fatal("the tutor bullet should have queued a search choice")
	}
	answerSearchNamed(t, g, me.ID, "basic-filler")

	if got := me.Library.Size(); got != libBefore {
		t.Errorf("library size after put-on-top = %d, want %d (nothing leaves the library)", got, libBefore)
	}
}

// The targeted draw-and-drain bullet alone.
func TestInsatiableAvariceDrawsThreeAndDrainsAlone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	handBefore := opp.Hand.Size()
	lifeBefore := opp.Life

	castModal(t, g, "Insatiable Avarice", "Sorcery", insatiableAvariceOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetPlayer, opp.ID, 0, 0)})
	passPriorityAroundTable(t, g)

	if got, want := opp.Hand.Size(), handBefore+3; got != want {
		t.Errorf("opponent's hand after the drain bullet = %d, want %d", got, want)
	}
	if got, want := opp.Life, lifeBefore-3; got != want {
		t.Errorf("opponent's life after the drain bullet = %d, want %d", got, want)
	}
}

// Both bullets (#2789, CR 608.2c): "search for a card, put it on top",
// THEN "target player draws three". The draw waits for the search, so
// casting it on yourself draws the card you tutored.
func TestInsatiableAvariceDrawsTheTutoredCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Tutored", TypeLine: "Sorcery"},
		game.Card{Name: "Filler A", TypeLine: "Sorcery"},
		game.Card{Name: "Filler B", TypeLine: "Sorcery"},
		game.Card{Name: "Filler C", TypeLine: "Sorcery"},
		game.Card{Name: "Filler D", TypeLine: "Sorcery"},
	)
	handBefore := me.Hand.Size()

	castModal(t, g, "Insatiable Avarice", "Sorcery", insatiableAvariceOracle,
		[]int{1, 0}, []game.TargetRef{modeRef(game.TargetPlayer, me.ID, 0, 0)})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Fatalf("the draw ran before the search was answered: hand %d → %d", handBefore, got)
	}
	answerSearchNamed(t, g, me.ID, "Tutored")
	g.SettleResolution()

	if got, want := me.Hand.Size(), handBefore+3; got != want {
		t.Errorf("hand after the draw = %d, want %d", got, want)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the tutored card was put on top before the draw, so it is drawn")
	}
}
