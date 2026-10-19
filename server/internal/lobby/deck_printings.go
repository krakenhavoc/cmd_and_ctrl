package lobby

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
)

type deckPrintingRow struct {
	Entry      int       `json:"entry"`
	Name       string    `json:"name"`
	Count      int       `json:"count"`
	Board      string    `json:"board"`
	PrintingID uuid.UUID `json:"printing_id"`
	Set        string    `json:"set"`
	Number     string    `json:"collector_number"`
}

type deckPrintingsResponse struct {
	Revision string            `json:"revision"`
	Cards    []deckPrintingRow `json:"cards"`
}

func printingRevision(d decklibrary.Deck) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(d.SourceFormat+"\x00"+d.SourceText)))
}

func printingBoard(e deck.Entry) string {
	switch {
	case e.IsCommander:
		return "Commander"
	case e.IsCompanion:
		return "Companion"
	case e.IsSideboard:
		return "Sideboard"
	default:
		return "Mainboard"
	}
}

// Moxfield JSON boards use maps. Sort before assigning row numbers so
// reading the same saved JSON twice produces the same selectable rows.
func printingEntries(d decklibrary.Deck) ([]deck.Entry, error) {
	entries, err := libraryDeckEntries(d)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool {
		key := func(e deck.Entry) string {
			return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%s\x00%d", printingBoard(e), e.Name, e.SetCode, e.CollectorNo, e.PrintingID, e.Count)
		}
		return key(entries[i]) < key(entries[j])
	})
	return entries, nil
}

func myDeckPrintings(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, err := libraryOwner(r)
	if err != nil {
		return err
	}
	d, err := ownedLibraryDeck(c, r, owner)
	if err != nil {
		return err
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "card index not loaded")
	}
	entries, err := printingEntries(d)
	if err != nil {
		return httpError(http.StatusUnprocessableEntity, "the saved deck could not be read")
	}
	out := deckPrintingsResponse{Revision: printingRevision(d), Cards: []deckPrintingRow{}}
	for i, e := range entries {
		row := deckPrintingRow{Entry: i, Name: e.Name, Count: e.Count, Board: printingBoard(e)}
		if card, ok := c.Cards.FindByName(e.Name); ok {
			if chosen, found := c.Cards.FindPrintingOf(card, e.PrintingID, e.SetCode, e.CollectorNo); found {
				card = chosen
			}
			row.PrintingID, row.Set, row.Number = card.ID, card.SetCode, card.CollectorNo
		}
		out.Cards = append(out.Cards, row)
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

func putMyDeckPrintings(c Config, w http.ResponseWriter, r *http.Request) error {
	owner, err := libraryOwner(r)
	if err != nil {
		return err
	}
	d, err := ownedLibraryDeck(c, r, owner)
	if err != nil {
		return err
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "card index not loaded")
	}
	var body struct {
		Revision string `json:"revision"`
		Choices  []struct {
			Entry      int       `json:"entry"`
			PrintingID uuid.UUID `json:"printing_id"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(&body); err != nil {
		return httpError(http.StatusBadRequest, "invalid printing choices")
	}
	if body.Revision != printingRevision(d) {
		return httpError(http.StatusConflict, "This deck changed. Reopen the printing picker and try again.")
	}
	if len(body.Choices) == 0 || len(body.Choices) > 2000 {
		return httpError(http.StatusBadRequest, "choose at least one printing")
	}
	entries, err := printingEntries(d)
	if err != nil {
		return httpError(http.StatusUnprocessableEntity, "the saved deck could not be read")
	}
	seen := map[int]bool{}
	for _, choice := range body.Choices {
		if choice.Entry < 0 || choice.Entry >= len(entries) || seen[choice.Entry] {
			return httpError(http.StatusBadRequest, "invalid or repeated card row")
		}
		seen[choice.Entry] = true
		e := &entries[choice.Entry]
		card, ok := c.Cards.FindByName(e.Name)
		if !ok {
			return httpError(http.StatusBadRequest, "that card could not be found in the card index")
		}
		selected, ok := c.Cards.FindPrintingOf(card, choice.PrintingID, "", "")
		if !ok || selected.ID != choice.PrintingID || selected.SetCode == "" || selected.CollectorNo == "" || (selected.Lang != "" && selected.Lang != "en") {
			return httpError(http.StatusBadRequest, "choose an available printing of the same card")
		}
		e.Name, e.PrintingID, e.SetCode, e.CollectorNo = selected.Name, selected.ID, selected.SetCode, selected.CollectorNo
	}
	// Render every row, including unresolved cards and distinct copies of
	// basics, with explicit boards. Changing art must never drop a card.
	var text strings.Builder
	for _, e := range entries {
		if e.PrintingID != uuid.Nil && (e.SetCode == "" || e.CollectorNo == "") {
			if card, ok := c.Cards.FindByName(e.Name); ok {
				if chosen, found := c.Cards.FindPrintingOf(card, e.PrintingID, e.SetCode, e.CollectorNo); found {
					e.SetCode, e.CollectorNo = chosen.SetCode, chosen.CollectorNo
				}
			}
		}
		fmt.Fprintf(&text, "%s:\n%d %s", printingBoard(e), e.Count, e.Name)
		if e.SetCode != "" && e.CollectorNo != "" {
			fmt.Fprintf(&text, " (%s) %s", strings.ToUpper(e.SetCode), e.CollectorNo)
		}
		text.WriteByte('\n')
	}
	updated, err := c.libraryStore().UpdatePrintings(r.Context(), d, text.String())
	if errors.Is(err, decklibrary.ErrSourceChanged) {
		return httpError(http.StatusConflict, err.Error())
	}
	if errors.Is(err, decklibrary.ErrNotFound) {
		return httpError(http.StatusNotFound, "deck not found")
	}
	if err != nil {
		return err
	}
	verdicts, err := deckcoverage.Verdicts(c.Cards)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, libraryDeckInfo(c, updated, verdicts))
}
