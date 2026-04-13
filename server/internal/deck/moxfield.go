package deck

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ParseMoxfield parses a Moxfield export (JSON) into []Entry. The
// Moxfield API emits at least two shapes depending on the export
// endpoint; the one we accept is the "v3 deck" shape produced by
// the "Export → JSON" button on any public deck page, which has
// the form:
//
//	{
//	  "name": "My Deck",
//	  "commanders": { "<card-name>": { "quantity": N, ... } },
//	  "mainboard":  { "<card-name>": { "quantity": N, ... } },
//	  "sideboard":  { ... },
//	  "companions": { ... }
//	}
//
// Card records carry more fields than we need (mana cost, colors,
// Scryfall ID, etc.); we read only quantity + the map key as the
// card name. If Moxfield ever ships a stable Scryfall UUID on each
// entry, switching to ID-based resolution is a one-field change.
func ParseMoxfield(raw []byte) (name string, entries []Entry, err error) {
	var doc struct {
		Name       string                   `json:"name"`
		Commanders map[string]moxfieldEntry `json:"commanders"`
		Mainboard  map[string]moxfieldEntry `json:"mainboard"`
		Sideboard  map[string]moxfieldEntry `json:"sideboard"`
		Companions map[string]moxfieldEntry `json:"companions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", nil, fmt.Errorf("deck: moxfield json: %w", err)
	}
	if doc.Commanders == nil && doc.Mainboard == nil {
		return "", nil, errors.New("deck: moxfield export had no commanders or mainboard — maybe the wrong JSON shape?")
	}

	// Companion slot is unsupported at S05 — Resolve would catch the
	// mechanic via oracle text anyway, but failing here gives a
	// clearer error message up front.
	if len(doc.Companions) > 0 {
		return "", nil, ErrUnsupportedMechanic
	}

	for n, e := range doc.Commanders {
		if e.Quantity <= 0 {
			continue
		}
		entries = append(entries, Entry{Name: n, Count: e.Quantity, IsCommander: true})
	}
	for n, e := range doc.Mainboard {
		if e.Quantity <= 0 {
			continue
		}
		entries = append(entries, Entry{Name: n, Count: e.Quantity})
	}
	for n, e := range doc.Sideboard {
		if e.Quantity <= 0 {
			continue
		}
		entries = append(entries, Entry{Name: n, Count: e.Quantity, IsSideboard: true})
	}
	return doc.Name, entries, nil
}

// moxfieldEntry is the trimmed per-card record shape. Moxfield ships
// much more (Scryfall ID, foil flag, condition, tags, board
// history) — we carry only quantity.
type moxfieldEntry struct {
	Quantity int `json:"quantity"`
}
