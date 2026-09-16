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

// archidektAPIHost is Archidekt's public deck API. Package-level
// var (not const) so tests can swap in an httptest.Server. Archidekt
// serves its API off the same hostname as the UI, differentiated by
// path ("/api/decks/<id>/").
var archidektAPIHost = "https://archidekt.com"

// archidektDeck mirrors the subset of the Archidekt deck export
// payload we consume. The real response is considerably richer
// (format info, tags, viewership, category definitions, etc.); we
// pull only what the importer needs.
//
// Category names in Archidekt are user-configurable, but the
// conventional set includes "Commander", "Sideboard", and
// "Maybeboard"; a card with no categories or an explicit
// "Mainboard" label goes to the mainboard. We match on the first
// recognised category so a card tagged "Commander,Legendary" (the
// user added their own "Legendary" tag) still lands in the command
// zone.
type archidektDeck struct {
	Name  string          `json:"name"`
	Cards []archidektCard `json:"cards"`
}

type archidektCard struct {
	Quantity   int      `json:"quantity"`
	Categories []string `json:"categories"`
	Card       struct {
		OracleCard struct {
			Name string `json:"name"`
		} `json:"oracleCard"`
	} `json:"card"`
}

// ParseArchidekt parses the Archidekt public deck JSON into the
// shared Entry shape. Exposed for symmetry with ParseMoxfield / for
// tests; in normal use it's reached via FetchFromURL.
func ParseArchidekt(raw []byte) (name string, entries []Entry, err error) {
	var doc archidektDeck
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", nil, fmt.Errorf("deck: archidekt json: %w", err)
	}
	if len(doc.Cards) == 0 {
		return "", nil, errors.New("deck: archidekt export had no cards — maybe the wrong JSON shape or a private deck?")
	}

	for _, c := range doc.Cards {
		if c.Quantity <= 0 {
			continue
		}
		name := c.Card.OracleCard.Name
		if name == "" {
			continue
		}
		entry := Entry{Name: name, Count: c.Quantity}
		switch archidektBoard(c.Categories) {
		case boardCommander:
			entry.IsCommander = true
		case boardSideboard:
			entry.IsSideboard = true
		case boardMaybeboard:
			// Maybeboard cards aren't part of the deck list; skip so
			// validation doesn't complain about card-count overflow.
			continue
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return "", nil, errors.New("deck: archidekt export had no playable cards after filtering maybeboard")
	}
	return doc.Name, entries, nil
}

// board discriminates the three Archidekt category families we act
// on. Anything else falls through to mainboard.
type board int

const (
	boardMainboard board = iota
	boardCommander
	boardSideboard
	boardMaybeboard
)

// archidektBoard picks the first recognised category on a card.
// User-defined categories are ignored — we only look for the three
// board-relevant names. Case-insensitive because Archidekt's UI
// preserves capitalisation as typed by the user but the common
// names are consistent in practice.
func archidektBoard(categories []string) board {
	for _, cat := range categories {
		switch strings.ToLower(strings.TrimSpace(cat)) {
		case "commander":
			return boardCommander
		case "sideboard":
			return boardSideboard
		case "maybeboard":
			return boardMaybeboard
		}
	}
	return boardMainboard
}

// fetchArchidekt fetches a public Archidekt deck by ID and hands
// the bytes off to ParseArchidekt. URL shape is /decks/<id> or
// /decks/<id>/<slug>, matching Archidekt's UI format.
func fetchArchidekt(ctx context.Context, client *http.Client, u *url.URL) (name string, entries []Entry, err error) {
	deckID, err := extractArchidektDeckID(u)
	if err != nil {
		return "", nil, err
	}
	reqURL := archidektAPIHost + "/api/decks/" + url.PathEscape(deckID) + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("%w: build request: %v", ErrExternalAPIUnavailable, err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("%w: archidekt fetch: %v", ErrExternalAPIUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if werr := classifyHTTPStatus(resp.StatusCode); werr != nil {
		return "", nil, fmt.Errorf("%w: archidekt", werr)
	}

	body, err := readLimitedBody(resp.Body, maxBodyBytes)
	if err != nil {
		return "", nil, fmt.Errorf("%w: read archidekt body: %v", ErrExternalAPIUnavailable, err)
	}
	return ParseArchidekt(body)
}

func extractArchidektDeckID(u *url.URL) (string, error) {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "decks" || parts[1] == "" {
		return "", fmt.Errorf("%w: archidekt URL path must be /decks/<id>, got %q", ErrUnknownSource, u.Path)
	}
	return parts[1], nil
}
