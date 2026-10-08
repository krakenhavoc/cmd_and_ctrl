package deck

// ascend_import_test.go — the INGESTION half of ascend (#2696, CR
// 702.131). Scryfall lists "Ascend" in `keywords` for the permanent and
// for the instant or sorcery; the importer stamps the canonical token,
// and the engine reads it from there, so an imported card earns the
// city's blessing with no catalog entry.

import (
	"os"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestAscendImportsOnAPermanentAndOnASpell(t *testing.T) {
	for _, c := range []cards.Card{
		{
			ID: uuid.New(), OracleID: uuid.New(), Name: "Spire Winder", Layout: "normal",
			TypeLine: "Creature — Snake", ManaCost: "{2}{U}", Power: "1", Toughness: "1",
			Keywords: []string{"Flying", "Ascend"},
			OracleText: "Flying\nAscend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
				"This creature gets +1/+1 as long as you have the city's blessing.",
		},
		{
			ID: uuid.New(), OracleID: uuid.New(), Name: "Secrets of the Golden City", Layout: "normal",
			TypeLine: "Sorcery", ManaCost: "{1}{U}{U}",
			Keywords: []string{"Ascend"},
			OracleText: "Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)\n" +
				"Draw two cards. If you have the city's blessing, draw three cards instead.",
		},
	} {
		got := toGameCard(c, false)
		if !slices.Contains(got.Keywords, game.KeywordAscend) {
			t.Errorf("%s: keywords = %v, want ascend stamped", c.Name, got.Keywords)
		}
		if !game.HasKeyword(&got, game.KeywordAscend) {
			t.Errorf("%s: HasKeyword(ascend) is false off the battlefield", c.Name)
		}
	}
}

// A card whose rules text merely mentions the city's blessing, with no
// ascend of its own (Scryfall's array is the only source), imports with
// no ascend: only a card that prints the keyword gets the static.
func TestACardThatOnlyReadsTheBlessingImportsWithoutAscend(t *testing.T) {
	c := cards.Card{
		ID: uuid.New(), OracleID: uuid.New(), Name: "Reader", Layout: "normal",
		TypeLine: "Creature — Test", ManaCost: "{2}{U}", Power: "1", Toughness: "1",
		Keywords:   []string{},
		OracleText: "This creature gets +1/+1 as long as you have the city's blessing.",
	}
	if got := toGameCard(c, false); slices.Contains(got.Keywords, game.KeywordAscend) {
		t.Errorf("keywords = %v, ascend must not be invented from the text", got.Keywords)
	}
}

// TestRealDumpAscendCardsImportWithTheKeyword is the real-dump half:
// every card whose Scryfall `keywords` list Ascend imports with it.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/deck/ -run RealDumpAscend -v
func TestRealDumpAscendCardsImportWithTheKeyword(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	seen := map[uuid.UUID]bool{}
	var missing []string
	streamDump(t, path, func(c cards.Card) {
		if !slices.Contains(c.Keywords, "Ascend") || seen[c.OracleID] {
			return
		}
		seen[c.OracleID] = true
		if got := toGameCard(c, false); !slices.Contains(got.Keywords, game.KeywordAscend) {
			missing = append(missing, got.Name)
		}
	})
	t.Logf("%d ascend oracle IDs", len(seen))
	if len(seen) == 0 {
		t.Fatal("no ascend card in the dump: the sweep proved nothing")
	}
	if len(missing) > 0 {
		t.Errorf("%d ascend cards import without the keyword: %v", len(missing), missing)
	}
}
