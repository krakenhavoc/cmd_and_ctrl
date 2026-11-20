package testsi

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The untargeted tutor bullet alone: a real search choice opens (the
// library is 20 identical matches, so nothing auto-resolves — the
// S22 "you choose" prompt this card shares with Vampiric Tutor), and
// answering it leaves the library the same SIZE — ToTop relocates
// the found card, it does not remove one.
func TestInsatiableAvariceTutorsToTopAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()

	effectstest.CastModal(t, g, "Insatiable Avarice", "Sorcery", effectstest.InsatiableAvariceOracle,
		[]int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil {
		t.Fatal("the tutor bullet should have queued a search choice")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "basic-filler")

	if got := me.Library.Size(); got != libBefore {
		t.Errorf("library size after put-on-top = %d, want %d (nothing leaves the library)", got, libBefore)
	}
}

// The targeted draw-and-drain bullet alone.
func TestInsatiableAvariceDrawsThreeAndDrainsAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	handBefore := opp.Hand.Size()
	lifeBefore := opp.Life

	effectstest.CastModal(t, g, "Insatiable Avarice", "Sorcery", effectstest.InsatiableAvariceOracle,
		[]int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Tutored", TypeLine: "Sorcery"},
		game.Card{Name: "Filler A", TypeLine: "Sorcery"},
		game.Card{Name: "Filler B", TypeLine: "Sorcery"},
		game.Card{Name: "Filler C", TypeLine: "Sorcery"},
		game.Card{Name: "Filler D", TypeLine: "Sorcery"},
	)
	handBefore := me.Hand.Size()

	effectstest.CastModal(t, g, "Insatiable Avarice", "Sorcery", effectstest.InsatiableAvariceOracle,
		[]int{1, 0}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore {
		t.Fatalf("the draw ran before the search was answered: hand %d → %d", handBefore, got)
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Tutored")
	g.SettleResolution()

	if got, want := me.Hand.Size(), handBefore+3; got != want {
		t.Errorf("hand after the draw = %d, want %d", got, want)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the tutored card was put on top before the draw, so it is drawn")
	}
}
