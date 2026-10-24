package deck

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_test.go — ADR 0208 §2: a deck's sticker sheets are not cards,
// not one of the 100, and reach the game as data when there are at least
// ten, each different (CR 123.2a).

const nightBrushwaggSheetText = "{TK}{TK} — Menace\n{TK}{TK}{TK} — Persist (When this permanent dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)\n{TK}{TK} — 2/3\n{TK}{TK}{TK}{TK}{TK}{TK} — 10/10"

// stickerSheetCards are n distinct legal sheet records.
func stickerSheetCards(n int) []cards.Card {
	out := make([]cards.Card, 0, n)
	for i := range n {
		c := basicLegal(fmt.Sprintf("Sheet%02d Word Word", i), StickerSheetTypeLine)
		c.OracleID = uuid.New()
		c.OracleText = nightBrushwaggSheetText
		out = append(out, c)
	}
	return out
}

// The importer reads a sheet's stickers off its record: three name
// stickers (one of them two words on Ancestral Hot Dog Minotaur), the
// ability and power/toughness stickers with their ticket costs, and three
// art stickers.
func TestStickerSheetOfReadsEveryKind(t *testing.T) {
	c := basicLegal("Night Brushwagg Ringmaster", StickerSheetTypeLine)
	c.OracleID = uuid.New()
	c.OracleText = nightBrushwaggSheetText
	sheet := StickerSheetOf(c)
	type row struct {
		kind  game.StickerKind
		text  string
		cost  int
		p, tg int
	}
	var got []row
	for _, s := range sheet.Stickers {
		got = append(got, row{s.Kind, s.Text, s.TicketCost, s.Power, s.Toughness})
	}
	want := []row{
		{game.StickerName, "Night", 0, 0, 0},
		{game.StickerName, "Brushwagg", 0, 0, 0},
		{game.StickerName, "Ringmaster", 0, 0, 0},
		{game.StickerAbility, "Menace", 2, 0, 0},
		{game.StickerAbility, "Persist", 3, 0, 0},
		{game.StickerPT, "2/3", 2, 2, 3},
		{game.StickerPT, "10/10", 6, 10, 10},
		{game.StickerArt, "", 0, 0, 0},
		{game.StickerArt, "", 0, 0, 0},
		{game.StickerArt, "", 0, 0, 0},
	}
	if !slices.Equal(got, want) {
		t.Errorf("stickers:\n got %v\nwant %v", got, want)
	}
	if sheet.Stickers[4].Ordinal != 1 || sheet.Stickers[6].Ordinal != 1 {
		t.Errorf("ordinals: %+v", sheet.Stickers)
	}
	hotDog := basicLegal("Ancestral Hot Dog Minotaur", StickerSheetTypeLine)
	var names []string
	for _, s := range StickerSheetOf(hotDog).Stickers {
		if s.Kind == game.StickerName {
			names = append(names, s.Text)
		}
	}
	if !slices.Equal(names, []string{"Ancestral", "Hot Dog", "Minotaur"}) {
		t.Errorf("Ancestral Hot Dog Minotaur's name stickers = %q", names)
	}
}

// Every importer routes a sheet by its type line, never into the 100.
func TestEveryImporterPutsSheetsWithTheSheets(t *testing.T) {
	cmd := basicLegal("Test Commander", "Legendary Creature — Test", "R")
	sheets := stickerSheetCards(10)
	idx := indexWith(append([]cards.Card{cmd, basicLegal("Lightning Bolt", "Instant", "R")}, sheets...)...)
	var rows strings.Builder
	for _, c := range sheets {
		fmt.Fprintf(&rows, "1 %s\n", c.Name)
	}
	for _, src := range []string{
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n\nStickers:\n" + rows.String(),
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n\nSticker sheets\n" + rows.String(),
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n" + rows.String(),
		"Commander:\n1 Test Commander\n\nDeck:\n1 Lightning Bolt\n\nSideboard:\n" + rows.String(),
	} {
		entries, err := ParseText(src)
		if err != nil {
			t.Fatal(err)
		}
		list, err := Resolve(idx, "t", entries)
		if err != nil {
			t.Fatal(err)
		}
		if len(list.StickerSheets) != 10 || len(list.Mainboard) != 1 || len(list.Sideboard) != 0 {
			t.Errorf("%q: %d sheets, %d main, %d side; want 10, 1, 0", src[len(src)-30:], len(list.StickerSheets), len(list.Mainboard), len(list.Sideboard))
		}
	}
}

// A legal pool rides List.Text and seats as data, not cards.
func TestALegalStickerPoolSeatsAsData(t *testing.T) {
	list := legalDeck()
	list.StickerSheets = stickerSheetCards(10)
	if err := Validate(list); err != nil {
		t.Fatalf("a legal deck with ten sheets: %v", err)
	}
	if text := list.Text(); !strings.Contains(text, "Stickers:\n") {
		t.Errorf("no Stickers section in the saved text")
	}
	out := list.ToGameCards()
	if len(out) != 110 {
		t.Fatalf("%d game cards, want 100 + 10 sheet pseudo-cards", len(out))
	}
	for _, c := range out[100:] {
		if c.StickerSheet == nil || len(c.StickerSheet.Stickers) != 10 {
			t.Errorf("%s: sheet %+v", c.Name, c.StickerSheet)
		}
	}
	p := seatAndCount(t, list)
	if len(p.StickerSheetPool) != 10 || p.Library.Size()+p.Command.Size() != 100 {
		t.Errorf("pool %d, library+command %d", len(p.StickerSheetPool), p.Library.Size()+p.Command.Size())
	}
}

// CR 123.2a: fewer than ten, or one twice, is said (non-fatal) and no
// sheet reaches the game; the deck still seats.
func TestAnIllegalStickerPoolIsSaidAndNotSeated(t *testing.T) {
	dup := stickerSheetCards(10)
	dup[9] = dup[0]
	for name, sheets := range map[string][]cards.Card{"nine": stickerSheetCards(9), "one twice": dup} {
		t.Run(name, func(t *testing.T) {
			list := legalDeck()
			list.StickerSheets = sheets
			codes := violationCodes(Validate(list))
			if !slices.Equal(codes, []string{CodeStickerSheetsIllegal}) {
				t.Fatalf("codes %v, want only %s", codes, CodeStickerSheetsIllegal)
			}
			if p := seatAndCount(t, list); len(p.StickerSheetPool) != 0 {
				t.Error("an illegal pool reached the game")
			}
		})
	}
}
