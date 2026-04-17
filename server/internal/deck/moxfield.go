package deck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

// moxfieldAPIHost is where our outbound deck-fetch requests go.
// Kept as a package-level var rather than a const so tests can swap
// in an httptest.Server URL without exercising the live network.
var moxfieldAPIHost = "https://api2.moxfield.com"

// fetchMoxfield fetches a public Moxfield deck by ID and hands the
// bytes off to ParseMoxfield. The URL path is parsed to extract the
// deck ID — Moxfield's public URLs have the form
// `https://moxfield.com/decks/<id>` or the deprecated
// `https://moxfield.com/decks/<id>/<slug>`. Either works.
func fetchMoxfield(ctx context.Context, client *http.Client, u *url.URL) (name string, entries []Entry, err error) {
	deckID, err := extractMoxfieldDeckID(u)
	if err != nil {
		return "", nil, err
	}

	reqURL := moxfieldAPIHost + "/v3/decks/all/" + url.PathEscape(deckID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("%w: build request: %v", ErrExternalAPIUnavailable, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("%w: moxfield fetch: %v", ErrExternalAPIUnavailable, err)
	}
	defer resp.Body.Close()

	if werr := classifyHTTPStatus(resp.StatusCode); werr != nil {
		return "", nil, fmt.Errorf("%w: moxfield", werr)
	}

	body, err := readLimitedBody(resp.Body, maxBodyBytes)
	if err != nil {
		return "", nil, fmt.Errorf("%w: read moxfield body: %v", ErrExternalAPIUnavailable, err)
	}
	return ParseMoxfield(body)
}

// extractMoxfieldDeckID pulls the deck ID out of a Moxfield URL path.
// Accepts the two forms Moxfield's UI produces:
//
//	/decks/<id>            — canonical short link
//	/decks/<id>/<slug>     — the full URL with a human-readable slug
//
// Anything else is surfaced as ErrUnknownSource so the caller can
// tell "wrong host / wrong path" apart from "upstream is down".
func extractMoxfieldDeckID(u *url.URL) (string, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "decks" || parts[1] == "" {
		return "", fmt.Errorf("%w: moxfield URL path must be /decks/<id>, got %q", ErrUnknownSource, u.Path)
	}
	return parts[1], nil
}
