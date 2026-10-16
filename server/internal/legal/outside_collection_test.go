package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// outside_collection_test.go — ADR 0198 §4: a pick from the card
// database is answerable by the enumerator (#544), from the prompt's own
// suggestions, and every answer it offers is one the engine accepts.

type legalCollection struct{ cards []game.Card }

func (f legalCollection) Search(string, int) []game.Card { return nil }
func (f legalCollection) Named(name string) (game.Card, bool) {
	for _, c := range f.cards {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return game.Card{}, false
}
func (f legalCollection) Suggest(fn func(game.Card) bool) {
	for _, c := range f.cards {
		if !fn(c) {
			return
		}
	}
}

func installLegalCollection(t *testing.T) {
	t.Helper()
	var cards []game.Card
	for _, n := range [][2]string{
		{"Alpha Sorcery", "Sorcery"}, {"Beta Sorcery", "Sorcery"}, {"Gamma Creature", "Creature — Bear"},
	} {
		c := game.NewCard(n[0], uuid.Nil)
		c.TypeLine = n[1]
		cards = append(cards, c)
	}
	prev := game.OutsideCollection
	game.OutsideCollection = legalCollection{cards: cards}
	t.Cleanup(func() { game.OutsideCollection = prev })
}

func TestACollectionPickIsEnumeratedAndEveryMoveIsAccepted(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t) // seat 0 has no sideboard
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: me.ID, Question: "Burning Wish — reveal a sorcery card",
			Match: func(c game.Card) bool { return c.IsSorcery() },
		})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	got := labelsOf(moves)
	for _, want := range []string{"take nothing", "take Alpha Sorcery", "take Beta Sorcery"} {
		if !strings.Contains(got, want) {
			t.Errorf("moves lack %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Gamma Creature") || len(moves) != 3 {
		t.Errorf("moves:\n%s", got)
	}
	always := 0
	for _, m := range moves {
		if m.AlwaysLegal {
			always++
		}
	}
	if always != 1 {
		t.Errorf("%d always-legal moves, want 1 (take nothing)", always)
	}
	dispatchAll(t, g, me.ID, moves)
	if opp := outsideChoiceMoves(legal.EnumerateFor(g, g.Seats[1].ID)); len(opp) != 0 {
		t.Errorf("the opponent was offered %d moves for my prompt", len(opp))
	}
}

func TestAMandatoryCollectionPickHasAnAnswer(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{Player: me.ID, Question: "Mastermind's Acquisition", Min: 1, Max: 1})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	if len(moves) != 3 || strings.Contains(labelsOf(moves), "take nothing") {
		t.Fatalf("moves:\n%s", labelsOf(moves))
	}
	dispatchAll(t, g, me.ID, moves)
}

func TestACollectionPairOffersOnlyPairsTheSetRuleAccepts(t *testing.T) {
	installLegalCollection(t)
	g := newSideboardTable(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOutsidePickForEffect(game.OutsidePickPrompt{
			Player: me.ID, Question: "Extrapolate the Impossible", Min: 0, Max: 2,
			Validate: func(picked []game.Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
		})
	})
	moves := outsideChoiceMoves(legal.EnumerateFor(g, me.ID))
	// take nothing, and the three pairs of three names; no single.
	if len(moves) != 4 {
		t.Fatalf("%d moves:\n%s", len(moves), labelsOf(moves))
	}
	for _, m := range moves {
		if !m.AlwaysLegal && strings.Count(m.Label, " and ") != 1 {
			t.Errorf("a set the rule refuses was offered: %q", m.Label)
		}
	}
	dispatchAll(t, g, me.ID, moves)
}
