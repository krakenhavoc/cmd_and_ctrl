package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// search_chooser_test.go — the S22 "you choose" search prompt, plus
// the #263 regression suite for the CR 614 entry pipeline the search
// path used to skip.
//
// Helpers first: nearly every tutor / fetch test in this package now
// has to answer a prompt, so the shapes live here rather than being
// re-spelled per file.

// searchChoiceFor returns the newest queued search prompt addressed
// to a player, or nil.
func SearchChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceSearchLibrary && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// searchOptionNamed returns the candidate with the given name from a
// search prompt, or uuid.Nil.
func SearchOptionNamed(g *game.Game, c *game.PendingChoice, name string) uuid.UUID {
	for _, id := range c.SearchCards {
		var found bool
		g.ReadSnapshot(func() {
			card, ok := g.LookupCardForEffect(id)
			found = ok && card.Name == name
		})
		if found {
			return id
		}
	}
	return uuid.Nil
}

// answerSearchNamed answers a player's open search prompt by taking
// the named cards. Fails the test when no prompt is open or a name
// is not among the candidates — "the prompt didn't appear" and "the
// prompt didn't offer the card" are different bugs and both want to
// be loud.
func AnswerSearchNamed(t *testing.T, g *game.Game, chooser uuid.UUID, names ...string) {
	t.Helper()
	c := SearchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	picks := make([]uuid.UUID, 0, len(names))
	for _, n := range names {
		id := SearchOptionNamed(g, c, n)
		if id == uuid.Nil {
			t.Fatalf("search prompt did not offer %q", n)
		}
		picks = append(picks, id)
	}
	BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
}

// answerSearchByID answers a player's open search prompt with
// specific instance IDs — for fixtures that seed two cards with the
// same name and need to tell them apart.
func AnswerSearchByID(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := SearchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
}

// answerSearchFailToFind answers an open search prompt by taking nothing (CR
// 701.23b).
func AnswerSearchFailToFind(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	c := SearchChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no search prompt for %s", chooser)
	}
	BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, chooser, nil); err != nil {
		t.Fatalf("ResolveSearchLibrary (fail to find): %v", err)
	}
}

// seedSearchLibrary replaces a player's library with exactly the
// named cards, bottom-first. Returns their instance IDs in the same
// order.
func SeedSearchLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, PushLibraryCardForTest(p, c))
	}
	return ids
}

func SearchTestLand(name, typeLine string) game.Card {
	return game.Card{Name: name, TypeLine: typeLine}
}

// searchFetchedCard finds a battlefield card by name.
func SearchFetchedCard(g *game.Game, name string) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out, ok = c, true
				return
			}
		}
	})
	return out, ok
}
