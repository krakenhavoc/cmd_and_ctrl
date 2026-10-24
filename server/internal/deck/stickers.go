package deck

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers.go — a deck's sticker sheets (CR 123.2, ADR 0208 §2).
//
// A sticker sheet is not a card (CR 123.2: "Sticker sheets are not cards
// and have no characteristics"), but Scryfall lists each one as a record
// with the type line "Stickers", and a deck list names it like a card.
// Resolve puts every such record in List.StickerSheets, wherever the list
// named it, as it does an Attraction (ADR 0207 §2): never one of the 100.
//
// CR 123.2a is the constructed rule, and Commander has none of its own
// (CR 903 says nothing about stickers): a player who plays with stickers
// starts with at least ten sheets, each unique. A pool that breaks it is
// warned (sticker_sheets_illegal, non-fatal) and none of it is seated,
// ADR 0191 §5's posture: the server never chooses which sheets to keep.

// MinStickerSheets is the fewest sheets a constructed player who plays
// with stickers starts with (CR 123.2a).
const MinStickerSheets = 10

// StickerSheetTypeLine is Scryfall's type line for a sticker sheet.
const StickerSheetTypeLine = "Stickers"

// PrintsStickerSheet reports whether a record is a sticker sheet.
func PrintsStickerSheet(c cards.Card) bool {
	return strings.TrimSpace(c.TypeLine) == StickerSheetTypeLine
}

// stickerSheetsLegal is CR 123.2a for a constructed pool.
func stickerSheetsLegal(sheets []cards.Card) bool {
	if len(sheets) < MinStickerSheets {
		return false
	}
	seen := make(map[string]bool, len(sheets))
	for _, c := range sheets {
		key := c.Name
		if c.OracleID != uuid.Nil {
			key = c.OracleID.String()
		}
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

// SeatedStickerSheets is the pool that reaches the game: the whole of it
// when it is legal (CR 123.2a), nothing otherwise.
func (l *List) SeatedStickerSheets() []cards.Card {
	if !stickerSheetsLegal(l.StickerSheets) {
		return nil
	}
	return l.StickerSheets
}

// stickerNameWords are the sheets whose name stickers are not one word
// each: Scryfall names a sheet by its name stickers' words, and a name
// sticker may carry more than one word (CR 123.6, "one or more words").
// Read off the printed sheets; every other legal sheet's three name
// stickers are its three words.
var stickerNameWords = map[string][]string{
	"Ancestral Hot Dog Minotaur": {"Ancestral", "Hot Dog", "Minotaur"},
}

// artStickersPerSheet is the number of art stickers on every Unfinity
// sticker sheet. Scryfall's text does not list them (they have no words
// and no cost, CR 123.9); the printed sheets each carry three.
const artStickersPerSheet = 3

var (
	ticketCostPrefix = regexp.MustCompile(`^((?:\{TK\})+)\s+—\s+(.*)$`)
	ptStickerText    = regexp.MustCompile(`^(\d+)/(\d+)$`)
	reminderText     = regexp.MustCompile(`\s*\([^)]*\)`)
)

// StickerSheetOf reads a sheet record into the game's StickerSheet: its
// name stickers (no ticket cost), its ability stickers and power and
// toughness stickers in printed order with their ticket costs, and its
// art stickers. Exported for the deck checker and for tests.
func StickerSheetOf(c cards.Card) game.StickerSheet {
	sheet := game.StickerSheet{Name: c.Name}
	if c.OracleID != uuid.Nil {
		sheet.OracleID = c.OracleID.String()
	}
	if c.ID != uuid.Nil {
		sheet.ScryfallID = c.ID.String()
	}
	words, ok := stickerNameWords[c.Name]
	if !ok {
		words = strings.Fields(c.Name)
	}
	for i, w := range words {
		sheet.Stickers = append(sheet.Stickers, game.StickerPrint{Kind: game.StickerName, Text: w, Ordinal: i})
	}
	abilities, pts := 0, 0
	for _, line := range strings.Split(c.OracleText, "\n") {
		m := ticketCostPrefix.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		cost := strings.Count(m[1], "{TK}")
		text := strings.TrimSpace(m[2])
		if pt := ptStickerText.FindStringSubmatch(text); pt != nil {
			p, _ := strconv.Atoi(pt[1])
			t, _ := strconv.Atoi(pt[2])
			sheet.Stickers = append(sheet.Stickers, game.StickerPrint{Kind: game.StickerPT, Text: text, TicketCost: cost, Power: p, Toughness: t, Ordinal: pts})
			pts++
			continue
		}
		sheet.Stickers = append(sheet.Stickers, game.StickerPrint{
			Kind:       game.StickerAbility,
			Text:       strings.TrimSpace(reminderText.ReplaceAllString(text, "")),
			TicketCost: cost,
			Ordinal:    abilities,
		})
		abilities++
	}
	for i := range artStickersPerSheet {
		sheet.Stickers = append(sheet.Stickers, game.StickerPrint{Kind: game.StickerArt, Ordinal: i})
	}
	return sheet
}

// stickerSheetCard is the pseudo-card that hands one sheet to the game
// (game.Card.StickerSheet): seated as data, never in a zone.
func stickerSheetCard(c cards.Card) game.Card {
	sheet := StickerSheetOf(c)
	out := game.Card{
		InstanceID: uuid.New(),
		Name:       c.Name,
		OracleID:   sheet.OracleID,
		ScryfallID: sheet.ScryfallID,
		TypeLine:   c.TypeLine,
	}
	out.StickerSheet = &sheet
	return out
}
