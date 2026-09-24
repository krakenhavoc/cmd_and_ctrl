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
