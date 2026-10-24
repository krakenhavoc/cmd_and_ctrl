package deckcoverage

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// ADR 0208 §2: a deck's sticker sheets are reported on their own, never
// among the 100, and bucketed by what the catalog runs.
func TestStickerSheetsAreReportedOutsideTheDeck(t *testing.T) {
	idx, tc := TestingIndex(t)
	base, err := Build(idx, Deck{Source: "text", Entries: testingDeck(tc)})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	legal := map[string]string{"commander": "legal"}
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.MustParse("1436370f-baf0-4659-804e-90aad871b05b"),
		Name: "Night Brushwagg Ringmaster", TypeLine: deck.StickerSheetTypeLine, Legalities: legal,
		OracleText: "{TK}{TK} — Menace\n{TK}{TK}{TK} — Persist\n{TK}{TK} — 2/3\n{TK}{TK}{TK}{TK}{TK}{TK} — 10/10"})
	idx.Put(cards.Card{ID: uuid.New(), OracleID: uuid.New(),
		Name: "Uncatalogued Test Sheet", TypeLine: deck.StickerSheetTypeLine, Legalities: legal,
		OracleText: "{TK}{TK} — {T}: Do something no catalog card does.\n{TK}{TK} — 2/2"})

	entries := append(testingDeck(tc),
		deck.Entry{Name: "Night Brushwagg Ringmaster", Count: 1},
		deck.Entry{Name: "Uncatalogued Test Sheet", Count: 1, IsSideboard: true},
	)
	r, err := Build(idx, Deck{Source: "text", Entries: entries})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	got := map[string]Bucket{}
	for _, c := range r.StickerSheets {
		got[c.Name] = c.Bucket
	}
	if got["Night Brushwagg Ringmaster"] != Automated || len(got) != 2 {
		t.Errorf("sticker sheets = %+v, want the catalogued sheet automated and both listed", r.StickerSheets)
	}
	if !reflect.DeepEqual(r.Cards, base.Cards) || !reflect.DeepEqual(r.Counts, base.Counts) || r.Sideboard != nil {
		t.Errorf("the sheets changed the deck's own report: cards %d vs %d, sideboard %+v", len(r.Cards), len(base.Cards), r.Sideboard)
	}
}
