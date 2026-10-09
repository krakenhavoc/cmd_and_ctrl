package lobby

// mydecks_save.go is ADR 0112 §3 items 4 and 5 (Delivery PR 2): the
// decks page's two writes that do not happen at a table.
//
//   - POST /me/decks saves a checked deck to the caller's library. A
//     pasted list is saved as pasted. A link is saved as the list the
//     server fetched, with the canonical link beside it, exactly as an
//     import at a table is (saveToLibrary). The check's cache keeps the
//     fetched list, so a save after a check fetches nothing.
//   - POST /deck-requests {deck_id} asks for a saved deck's missing
//     cards. savedDeckForRequest reads the deck and keys it the way its
//     original would have been keyed.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
)

// saveDeckRequest is the body of POST /me/decks: url or text, exactly
// one, and an optional name.
type saveDeckRequest struct {
	URL  string `json:"url,omitempty"`
	Text string `json:"text,omitempty"`
	Name string `json:"name,omitempty"`
}

// saveDeckResponse is the body of POST /me/decks. Deck is the GET
// /me/decks entry, coverage included. Replaced is true when the caller
// already had a deck by that name and it was replaced.
type saveDeckResponse struct {
	Deck     myDeckInfo `json:"deck"`
	Replaced bool       `json:"replaced"`
}

// libraryFullCode is the 409's code when a new deck would pass the cap.
const libraryFullCode = "library_full"

// fetchLimitRetryAfter is the Retry-After a save refused by the fetch
// limit sends: POST /deck-coverage's bucket refills one token every ten
// seconds.
const fetchLimitRetryAfter = 10

// saveDeck handles POST /me/decks (ADR 0112 §3 item 4).
//
// The caller is a signed-in person, else 403 (never 401: authFetch
// signs the browser out). The route rides the /me/decks* bucket. A
// link the check's cache does not hold spends one token from POST
// /deck-coverage's per-IP bucket before it is fetched, so saving
// cannot get around the fetch limit; a cached link and a pasted list
// spend none.
//
// A check writes nothing, and a save is always this explicit call
// (owner answer 2). A save never files a request either: the two are
// separate buttons with separate limits (§3 item 6).
//
// The deck is not validated against Commander rules here, and names
// the card index cannot resolve are kept: the deck check reports both,
// the library's coverage counts the unknown names, and seating a saved
// deck re-runs the whole parse, resolve and validate pipeline against
// the catalog of that moment (seatLibraryDeck), as it always has.
func (d *deckCheck) saveDeck(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	var body saveDeckRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.URL, body.Text = strings.TrimSpace(body.URL), strings.TrimSpace(body.Text)
	switch {
	case body.URL == "" && body.Text == "":
		return httpError(http.StatusBadRequest, "send a deck link as url, or a pasted list as text")
	case body.URL != "" && body.Text != "":
		return httpError(http.StatusBadRequest, "send url or text, not both")
	}
	name := strings.TrimSpace(body.Name)
	if utf8.RuneCountInString(name) > maxDeckNameLen {
		return httpError(http.StatusBadRequest, fmt.Sprintf("a deck name is at most %d characters", maxDeckNameLen))
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}
	if c.DeckLibrary == nil {
		return httpError(http.StatusServiceUnavailable, "deck library not configured")
	}

	var (
		entries    []deck.Entry
		sourceText string
		sourceURL  string
		fetched    string // the deck site's name for a link
	)
	if body.URL != "" {
		spend := func() bool { return d.fetchLimit == nil || d.fetchLimit.AllowRequest(r) }
		checked, err := d.deckForURL(r.Context(), c, body.URL, spend)
		if errors.Is(err, errFetchLimited) {
			w.Header().Set("Retry-After", strconv.Itoa(fetchLimitRetryAfter))
			return writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many deck checks from here. Wait a few seconds and save again.",
			})
		}
		if err != nil {
			return writeDeckFetchError(w, err)
		}
		// deckForURL already parsed the link without error.
		ref, _ := deck.ParseDeckURL(body.URL)
		entries, sourceText, sourceURL, fetched = checked.entries, entriesText(checked.entries), ref.URL(), checked.name
	} else {
		list, err := parseList(c, body.Text)
		if err != nil {
			return err
		}
		entries, sourceText = list.entries, body.Text
	}

	commanders, cardCount := libraryShape(c, entries)
	if name == "" {
		name = strings.TrimSpace(fetched)
	}
	if name == "" && len(commanders) > 0 {
		// A "Partner with" pair (#2142) is named for both commanders,
		// as libraryFallbackName names it.
		name = strings.Join(commanders, " / ")
	}
	if name == "" {
		name = "Untitled deck"
	}
	if utf8.RuneCountInString(name) > maxDeckNameLen {
		// A deck site's name is the caller's own text, but not one they
		// typed here: cut it rather than refuse the save.
		name = strings.TrimSpace(string([]rune(name)[:maxDeckNameLen]))
	}

	saved, replaced, err := c.DeckLibrary.Save(r.Context(), p.UserID, name, "text", sourceText, sourceURL, commanders, cardCount)
	switch {
	case errors.Is(err, decklibrary.ErrLibraryFull):
		return writeJSON(w, http.StatusConflict, map[string]string{
			"error": fmt.Sprintf("Your deck library is full (%d decks). Delete one below to save this one.", decklibrary.MaxDecks),
			"code":  libraryFullCode,
		})
	case err != nil:
		c.logger().Error("decklibrary: saving a checked deck failed", "err", err, "user_id", p.UserID)
		return httpError(http.StatusInternalServerError, "could not save your deck; try again")
	}

	info := libraryDeckInfo(c, saved, libraryVerdicts(c))
	status := http.StatusCreated
	if replaced {
		status = http.StatusOK
	}
	return writeJSON(w, status, saveDeckResponse{Deck: info, Replaced: replaced})
}

// entriesText renders a fetched deck as a plain-text list in sections,
// the form a saved link deck keeps (ADR 0110 owner decision 7):
// re-seating it, or requesting its cards, re-parses this text and never
// calls the network.
func entriesText(entries []deck.Entry) string {
	var b strings.Builder
	section := func(header string, keep func(deck.Entry) bool) {
		wrote := false
		for _, e := range entries {
			if e.Count <= 0 || !keep(e) {
				continue
			}
			if !wrote {
				b.WriteString(header + "\n")
				wrote = true
			}
			fmt.Fprintf(&b, "%d %s\n", e.Count, e.Name)
		}
	}
	section("Commander:", func(e deck.Entry) bool { return e.IsCommander })
	section("Mainboard:", func(e deck.Entry) bool { return !e.IsCommander && !e.IsSideboard })
	section("Sideboard:", func(e deck.Entry) bool { return !e.IsCommander && e.IsSideboard })
	return b.String()
}

// libraryShape is a list's commanders (one per copy, under the card
// index's name when it knows the card) and its card count without the
// sideboard: the two columns a library row keeps beside the list, as
// saveToLibrary fills them from a resolved deck.
func libraryShape(c Config, entries []deck.Entry) ([]string, int) {
	commanders := []string{}
	count := 0
	for _, e := range entries {
		if e.Count <= 0 || (e.IsSideboard && !e.IsCommander) {
			continue
		}
		count += e.Count
		if !e.IsCommander {
			continue
		}
		name := strings.TrimSpace(e.Name)
		if c.Cards != nil {
			if card, ok := c.Cards.FindByName(name); ok {
				name = card.Name
			}
		}
		for range e.Count {
			commanders = append(commanders, name)
		}
	}
	return commanders, count
}

// savedDeckRequest is a saved deck read for POST /deck-requests
// {deck_id}: its stored list, and the key it files under.
type savedDeckRequest struct {
	deck    decklibrary.Deck
	entries []deck.Entry
	// link is the deck's source link, nil for a pasted deck (or a link
	// that no longer parses as a deck link).
	link *deck.SourceRef
	key  string
	// list is the keyed stored list, for a deck without a link.
	list parsedList
}

// savedDeckForRequest reads one of owner's saved decks for a request
// (ADR 0112 §3 item 5). A deck that is not theirs, or no deck at all,
// is a 404, as on the other /me/decks routes. The stored list is
// re-parsed, never fetched, and keyed the way the original would have
// been: a deck with a source_url keys as that link, so it joins the
// issue a request from the link filed; any other deck keys as its list.
func savedDeckForRequest(ctx context.Context, c Config, owner uuid.UUID, rawID string) (*savedDeckRequest, error) {
	d, err := ownedLibraryDeckByID(ctx, c, owner, rawID)
	if err != nil {
		return nil, err
	}
	entries, err := libraryDeckEntries(d)
	if err != nil {
		return nil, httpError(http.StatusUnprocessableEntity, "that saved deck could not be read: "+err.Error())
	}
	out := &savedDeckRequest{deck: d, entries: entries}
	if d.SourceURL != "" {
		if ref, err := deck.ParseDeckURL(d.SourceURL); err == nil {
			out.link, out.key = &ref, ref.Key()
			return out, nil
		}
	}
	list, err := keyList(c, entries)
	if err != nil {
		return nil, err
	}
	out.list, out.key = list, list.key
	return out, nil
}

// libraryVerdicts builds the catalogue's verdicts once for a request
// that reports on saved decks, or nil without a card index.
func libraryVerdicts(c Config) map[string]catalog.Entry {
	if c.Cards == nil {
		return nil
	}
	v, _ := deckcoverage.Verdicts(c.Cards)
	return v
}
