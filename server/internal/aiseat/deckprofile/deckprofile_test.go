package deckprofile

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// Build is the static half of the model prompt. It has to
// survive having no Scryfall dump — CI has none — because a server
// that cannot load the index has a louder problem to report than a
// thin prompt.
func TestBuildWithoutACardIndex(t *testing.T) {
	for _, d := range decks.All() {
		p, ok := Build(nil, d.ID)
		if !ok {
			t.Fatalf("%s: no profile", d.ID)
		}
		if p.Name != d.Name {
			t.Errorf("%s: name = %q", d.ID, p.Name)
		}
		if !strings.Contains(p.Archetype, d.Archetype) {
			t.Errorf("%s: plan does not name the archetype: %q", d.ID, p.Archetype)
		}
		if !strings.Contains(p.Archetype, d.Commander.Name) {
			t.Errorf("%s: plan does not name the commander: %q", d.ID, p.Archetype)
		}
		// One entry per distinct card, commander included, and no
		// duplicate rows to pay tokens for twice.
		seen := map[string]bool{}
		for _, c := range p.Cards {
			if seen[c.Name] {
				t.Errorf("%s: %q appears twice in the profile", d.ID, c.Name)
			}
			seen[c.Name] = true
		}
		if !seen[d.Commander.Name] {
			t.Errorf("%s: the commander is missing from the profile", d.ID)
		}
		if len(p.Cards) < 50 {
			t.Errorf("%s: profile has %d cards, which is not a decklist", d.ID, len(p.Cards))
		}
	}

	if _, ok := Build(nil, "no-such-deck"); ok {
		t.Error("Build accepted an unknown deck ID")
	}
}

// With an index, the oracle text travels — that is the whole reason
// the prompt carries a decklist rather than a list of names.
func TestBuildTakesOracleTextFromTheIndex(t *testing.T) {
	d := decks.All()[0]
	idx := cards.NewIndex()
	idx.Put(cards.Card{
		ID:         uuid.New(),
		Name:       d.Commander.Name,
		ManaCost:   "{1}{U}{R}",
		TypeLine:   "Legendary Creature — Human Pirate",
		OracleText: "Whenever this creature deals combat damage, do a thing.",
	})

	p, ok := Build(idx, d.ID)
	if !ok {
		t.Fatal("no profile")
	}
	var found bool
	for _, c := range p.Cards {
		if c.Name != d.Commander.Name {
			continue
		}
		found = true
		if c.Cost != "{1}{U}{R}" || !strings.Contains(c.Oracle, "do a thing") || c.Type == "" {
			t.Errorf("commander entry = %+v", c)
		}
	}
	if !found {
		t.Fatal("the commander is not in the profile")
	}
	// A card the index does not have degrades to its name, not to an
	// error: a thinner prompt is a worse bot and not a broken one.
	for _, c := range p.Cards {
		if c.Name == "" {
			t.Error("a profile entry has no name")
		}
	}
}
