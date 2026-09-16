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

// ParseMoxfield parses a Moxfield v3 deck JSON into []Entry. The
// shape produced by https://api2.moxfield.com/v3/decks/all/<id> nests
// every board under a `boards` object, and each board carries
// `count` + `cards` where the inner keys are opaque Moxfield IDs
// (not the card names); the card name lives at `.card.name`:
//
//	{
//	  "name": "My Deck",
//	  "boards": {
//	    "commanders": { "count": N, "cards": { "<opaque>": { "quantity": N, "card": { "name": "..." } } } },
//	    "mainboard":  { ... },
//	    "sideboard":  { ... },
//	    "companions": { ... }
//	  }
//	}
//
// Card records carry more fields than we need (mana cost, colors,
// Scryfall ID, etc.); we read only quantity + card.name. Switching
// to Scryfall-UUID-based resolution is a later change — see the
// `card.scryfall_id` field in the response.
func ParseMoxfield(raw []byte) (name string, entries []Entry, err error) {
	var doc struct {
		Name   string `json:"name"`
		Boards struct {
			Mainboard  moxfieldBoard `json:"mainboard"`
			Sideboard  moxfieldBoard `json:"sideboard"`
			Commanders moxfieldBoard `json:"commanders"`
			Companions moxfieldBoard `json:"companions"`
		} `json:"boards"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return "", nil, fmt.Errorf("deck: moxfield json: %w", err)
	}
	if doc.Boards.Commanders.Cards == nil && doc.Boards.Mainboard.Cards == nil {
		return "", nil, errors.New("deck: moxfield export had no commanders or mainboard — maybe the wrong JSON shape?")
	}

	// Companion slot is unsupported at S05 — Resolve would catch the
	// mechanic via oracle text anyway, but failing here gives a
	// clearer error message up front.
	if len(doc.Boards.Companions.Cards) > 0 {
		return "", nil, ErrUnsupportedMechanic
	}

	entries = append(entries, collectMoxfieldBoard(doc.Boards.Commanders, func(e *Entry) { e.IsCommander = true })...)
	entries = append(entries, collectMoxfieldBoard(doc.Boards.Mainboard, nil)...)
	entries = append(entries, collectMoxfieldBoard(doc.Boards.Sideboard, func(e *Entry) { e.IsSideboard = true })...)
	return doc.Name, entries, nil
}

// moxfieldBoard is one of {mainboard, sideboard, commanders,
// companions}. The inner map is keyed by Moxfield's opaque per-card
// ID; the card identity we care about lives on .card.name.
type moxfieldBoard struct {
	Cards map[string]moxfieldBoardEntry `json:"cards"`
}

// moxfieldBoardEntry is the trimmed per-card record. Moxfield ships
// much more (foil flag, condition, tags, board history, Scryfall ID,
// etc.) — we carry only quantity + the nested card name.
type moxfieldBoardEntry struct {
	Quantity int `json:"quantity"`
	Card     struct {
		Name string `json:"name"`
	} `json:"card"`
}

// collectMoxfieldBoard walks a board and emits one Entry per card.
// tag, when non-nil, is invoked on each entry before append so the
// caller can stamp IsCommander / IsSideboard without a per-board loop
// duplicated three times.
func collectMoxfieldBoard(b moxfieldBoard, tag func(*Entry)) []Entry {
	if len(b.Cards) == 0 {
		return nil
	}
	out := make([]Entry, 0, len(b.Cards))
	for _, e := range b.Cards {
		if e.Quantity <= 0 || e.Card.Name == "" {
			continue
		}
		entry := Entry{Name: e.Card.Name, Count: e.Quantity}
		if tag != nil {
			tag(&entry)
		}
		out = append(out, entry)
	}
	return out
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
	defer func() { _ = resp.Body.Close() }()

	if werr := classifyHTTPStatus(resp.StatusCode); werr != nil {
		// 401/403 can be either a genuine private-deck response from
		// Moxfield (JSON body) or a Cloudflare bot-wall challenge
		// (HTML body). Re-classify the latter so the user sees a
		// useful error instead of "deck is private" when their deck
		// is actually public.
		if errors.Is(werr, ErrDeckPrivate) && looksLikeCloudflareBlock(resp) {
			return "", nil, fmt.Errorf("%w: moxfield (try Archidekt or paste the text export)", ErrUpstreamBlocked)
		}
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
