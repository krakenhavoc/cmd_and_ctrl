package deck

// semicolon_keyword_lines_test.go — #5791. Older Oracle wording puts
// several keywords on one line separated by semicolons rather than
// commas ("Flying; trample; rampage 4"), and the importer's line scan
// reads them the same way. The texts are the cards' real Oracle text
// from Scryfall, keeping only the fields the importer reads.

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	teekasDragonText      = "Flying; trample; rampage 4 (Whenever this creature becomes blocked, it gets +4/+4 until end of turn for each creature blocking it beyond the first.)"
	gorillaBerserkersText = "Trample; rampage 2 (Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it beyond the first.)\n" +
		"This creature can't be blocked except by three or more creatures."
)

// A semicolon-separated keyword line stamps every keyword on it, the
// numbered rampage included: Scryfall's array says only "Rampage", so
// its number comes from the line scan or not at all.
func TestImportReadsSemicolonKeywordLines(t *testing.T) {
	record := func(name, oracleID, typeLine string, keywords []string, text string) cards.Card {
		return cards.Card{
			ID: uuid.New(), OracleID: uuid.MustParse(oracleID), Name: name, Layout: "normal",
			TypeLine: typeLine, Power: "5", Toughness: "5", Keywords: keywords, OracleText: text,
		}
	}
	list := &List{Mainboard: []cards.Card{
		record("Teeka's Dragon", "6fb6e37e-d774-4a50-8ea3-b955535c2065", "Artifact Creature — Dragon",
			[]string{"Flying", "Trample", "Rampage"}, teekasDragonText),
		record("Gorilla Berserkers", "6373c66d-edfc-44cb-af2a-2273e0ba53e0", "Creature — Ape Berserker",
			[]string{"Trample", "Rampage"}, gorillaBerserkersText),
	}}
	got := map[string]game.Card{}
	for _, c := range list.ToGameCards() {
		got[c.Name] = c
	}
	for _, tc := range []struct {
		name    string
		want    []string
		rampage int
	}{
		{"Teeka's Dragon", []string{"flying", "trample", "rampage 4"}, 4},
		{"Gorilla Berserkers", []string{"trample", "rampage 2"}, 2},
	} {
		c, ok := got[tc.name]
		if !ok {
			t.Fatalf("%s missing from the imported deck", tc.name)
		}
		if len(c.Keywords) != len(tc.want) || !sameSet(c.Keywords, tc.want) {
			t.Errorf("%s: keywords = %v, want %v", tc.name, c.Keywords, tc.want)
		}
		if amounts := game.RampageAmounts(&c); len(amounts) != 1 || amounts[0] != tc.rampage {
			t.Errorf("RampageAmounts(imported %s) = %v, want [%d]", tc.name, amounts, tc.rampage)
		}
	}
}

// The line scan's counts for the same two lines: one of each keyword.
func TestKeywordLineCountsSplitsOnSemicolons(t *testing.T) {
	for text, want := range map[string]map[string]int{
		teekasDragonText:      {"flying": 1, "trample": 1, "rampage 4": 1},
		gorillaBerserkersText: {"trample": 1, "rampage 2": 1},
	} {
		got := keywordLineCounts(text)
		if len(got) != len(want) {
			t.Errorf("keywordLineCounts(%q) = %v, want %v", text, got, want)
			continue
		}
		for kw, n := range want {
			if got[kw] != n {
				t.Errorf("keywordLineCounts(%q)[%q] = %d, want %d", text, kw, got[kw], n)
			}
		}
	}
}

// A SENTENCE can use semicolons to list choices, keywords among them,
// and none of those is a keyword the card has. Built Bear's line has no
// full stop and three items that are canonical keywords on their own
// (deathtouch, reach, vigilance); Catch of the Day's ends with one and
// lists vigilance and islandwalk; Noxious Hydra Breath is a modal spell.
// All real Oracle text from Scryfall.
func TestSemicolonSentencesStampNoKeywords(t *testing.T) {
	for name, text := range map[string]string{
		"Built Bear": "Legacy (During deckbuilding, circle any number of the following. You can spend 1 point for free. For every 2 points spent beyond that, increase this card's mana cost by {1}.)\n" +
			"1 point each → Flash; deathtouch; reach; vigilance; or ward {2}\n" +
			"2 points → {T}: Add one mana of any color.\n" +
			"3 points each → +1/+1; or +0/+2\n" +
			"4 points → When this creature enters the battlefield, draw a card.",
		"Catch of the Day": "As Catch of the Day enters the battlefield, choose one for each of the following.\n" +
			"Keyword → Vigilance; ward {3}; or islandwalk.\n" +
			"Whenever this creature attacks → Scry 2; goad target creature an opponent controls; or tap target creature an opponent controls.\n" +
			"Size → 6/2; 4/4; or 2/6.",
		"Noxious Hydra Breath": "Choose one — Noxious Hydra Breath deals 5 damage to each player; or destroy each tapped non-Head creature.",
	} {
		if got := keywordLineCounts(text); len(got) != 0 {
			t.Errorf("%s: keywordLineCounts = %v, want none", name, got)
		}
	}
}
