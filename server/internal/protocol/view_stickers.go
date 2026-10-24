package protocol

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// view_stickers.go — stickers and sticker sheets on the wire (ADR 0208
// §6). Tickets need nothing here: a ticket counter is a player counter and
// rides PlayerView.Counters ("ticket"), as energy and rad do.

// StickerView is one sticker: on an object (CardView.Stickers) or on a
// sheet (StickerSheetView.Stickers).
type StickerView struct {
	// Kind is "name", "ability", "pt" or "art" (CR 123.1).
	Kind string `json:"kind"`
	// Text is what it says: a name sticker's word or words, an ability
	// sticker's ability, a power and toughness sticker's "2/3". Empty
	// for an art sticker.
	Text string `json:"text,omitempty"`
	// TicketCost is the number in its ticket symbol (CR 123.3c).
	TicketCost int `json:"ticket_cost,omitempty"`
	// Sheet is the name of the sheet it came from.
	Sheet string `json:"sheet,omitempty"`
	// InUse is set on a sheet's sticker that is on an object now, which
	// can't be put on another (CR 123.3). Never set on CardView.Stickers.
	InUse bool `json:"in_use,omitempty"`
}

// StickerSheetView is one sticker sheet a seat plays with.
type StickerSheetView struct {
	Name       string        `json:"name"`
	ScryfallID string        `json:"scryfall_id,omitempty"`
	Stickers   []StickerView `json:"stickers"`
}

// viewOfStickers is the stickers on an object, oldest first.
func viewOfStickers(c *game.Card) []StickerView {
	placed := game.StickersOf(c)
	if len(placed) == 0 {
		return nil
	}
	out := make([]StickerView, 0, len(placed))
	for _, s := range placed {
		out = append(out, StickerView{
			Kind:       string(s.Print.Kind),
			Text:       s.Print.Text,
			TicketCost: s.Print.TicketCost,
			Sheet:      s.SheetName,
		})
	}
	return out
}

// viewOfStickerSheets is the sheets a seat plays with, each sticker
// marked when it is on an object. Caller holds the game's read lock.
func viewOfStickerSheets(g *game.Game, p *game.Player) []StickerSheetView {
	sheets := game.StickerSheetsOf(p)
	if len(sheets) == 0 {
		return nil
	}
	used := g.StickersInUseLocked(p.ID)
	out := make([]StickerSheetView, 0, len(sheets))
	for _, sheet := range sheets {
		v := StickerSheetView{Name: sheet.Name, ScryfallID: sheet.ScryfallID, Stickers: make([]StickerView, 0, len(sheet.Stickers))}
		for i, s := range sheet.Stickers {
			v.Stickers = append(v.Stickers, StickerView{
				Kind:       string(s.Kind),
				Text:       s.Text,
				TicketCost: s.TicketCost,
				InUse:      used[game.StickerID(sheet.OracleID, i)],
			})
		}
		out = append(out, v)
	}
	return out
}
