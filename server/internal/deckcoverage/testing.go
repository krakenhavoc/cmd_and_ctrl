package deckcoverage

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
)

// TestingCards names one card per bucket in the index TestingIndex
// builds. The three catalogued ones are real registry entries, chosen
// at run time by their declared Completeness, so a card being audited
// from unreviewed to full never breaks a test that only needs "some
// full card".
type TestingCards struct {
	Commander  string // legendary vanilla creature: no_effect
	Vanilla    string // no_effect
	Basic      string // Forest: no_effect
	Manual     string // prints rules, not in the catalogue
	Automated  string
	Caveated   string
	Unreviewed string
	// CaveatText is the Caveated card's first caveat sentence.
	CaveatText string
}

// TestingIndex returns a small card index holding the TestingCards.
// Exported so the lobby's HTTP tests can build the same deck; like
// deck.TestingSetMoxfieldAPIHost, it takes a *testing.T because it
// has no production caller.
func TestingIndex(t *testing.T) (*cards.Index, TestingCards) {
	t.Helper()
	idx := cards.NewIndex()
	legal := map[string]string{"commander": "legal"}
	put := func(c cards.Card) {
		if c.ID == uuid.Nil {
			c.ID = uuid.New()
		}
		if c.Legalities == nil {
			c.Legalities = legal
		}
		idx.Put(c)
	}

	tc := TestingCards{
		Commander: "Testing Commander",
		Vanilla:   "Testing Bear",
		Basic:     "Forest",
		Manual:    "Testing Uncatalogued Engine",
	}
	put(cards.Card{OracleID: uuid.New(), Name: tc.Commander, TypeLine: "Legendary Creature — Human Warrior",
		Power: "2", Toughness: "2", ColorIdentity: []string{"G"}})
	put(cards.Card{OracleID: uuid.New(), Name: tc.Vanilla, TypeLine: "Creature — Bear", Power: "2", Toughness: "2",
		ColorIdentity: []string{"G"}})
	put(cards.Card{OracleID: uuid.New(), Name: tc.Basic, TypeLine: "Basic Land — Forest",
		OracleText: "({T}: Add {G}.)", ColorIdentity: []string{"G"}})
	put(cards.Card{OracleID: uuid.New(), Name: tc.Manual, TypeLine: "Artifact",
		OracleText: "At the beginning of your upkeep, do something no catalog card does."})

	pick := func(want effects.Completeness) catalog.Entry {
		t.Helper()
		var found []catalog.Entry
		for _, e := range catalog.Build(nil).Cards {
			if e.Completeness != want.String() || !effects.Has(e.OracleID) || effects.Has(e.OracleID+"#1") {
				continue
			}
			if _, err := uuid.Parse(e.OracleID); err != nil || e.Name == "" {
				continue
			}
			found = append(found, e)
		}
		if len(found) == 0 {
			t.Fatalf("the catalogue has no single-faced %s card to test with", want)
		}
		sort.Slice(found, func(a, b int) bool { return found[a].Name < found[b].Name })
		return found[0]
	}
	for _, p := range []struct {
		c    effects.Completeness
		name *string
	}{
		{effects.CompletenessFull, &tc.Automated},
		{effects.CompletenessCaveats, &tc.Caveated},
		{effects.CompletenessUnreviewed, &tc.Unreviewed},
	} {
		e := pick(p.c)
		*p.name = e.Name
		if p.c == effects.CompletenessCaveats {
			tc.CaveatText = e.Caveats[0]
		}
		// The oracle text only has to be rules text, so that the
		// card needs a catalogue entry; the registry keys on the
		// oracle ID, which is the real one.
		put(cards.Card{OracleID: uuid.MustParse(e.OracleID), Name: e.Name, TypeLine: "Artifact",
			OracleText: "{T}: Do what this card does."})
	}
	return idx, tc
}
