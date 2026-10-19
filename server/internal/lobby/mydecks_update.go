package lobby

// mydecks_update.go is #2036: updating a saved deck's contents from its
// own link, a new link or a pasted list (ADR 0095, amendment
// 2026-10-19).
//
//   - POST /me/decks/{id}/update is the PREVIEW. It reads the new list
//     the way the first import does, writes nothing, and answers with
//     the diff, the new list's coverage and Commander violations, and
//     what a request for its new missing cards would add. The new list
//     is held in memory under a one-use preview id.
//   - POST /me/decks/{id}/update/confirm {preview_id, request_missing}
//     COMMITS what was previewed: the previewed list, not a second
//     fetch, so what the player saw is what is saved. The request, when
//     asked for, is filed after the commit and never fails it.
//
// The deck keeps its id, its name and the printing choices of cards that
// carry over (carryPrintings). A table already seated with the deck is
// not touched: a seat holds the cards it was seated with (SetDeck copies
// them) and nothing re-reads the library row once a game exists.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckcoverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decklibrary"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckrequests"
)

const (
	// updatePreviewTTL is how long a previewed update waits to be
	// confirmed. Past it the confirm is a 410 and the player previews
	// again.
	updatePreviewTTL = 15 * time.Minute
	// updatePreviewMax bounds the previews held. A preview is the new
	// list and its report, tens of KiB; one person holds one per deck.
	updatePreviewMax = 128
	// maxUpdateURLLen bounds a link in an update body. A deck link is
	// short; the cap keeps a pathological one out of the parser.
	maxUpdateURLLen = 2048

	// updateRelink is the `source` value that re-pulls a deck from the
	// link it already has.
	updateRelink = "relink"
)

// updateDeckRequest is the body of POST /me/decks/{id}/update: exactly
// one of source ("relink"), url or list.
type updateDeckRequest struct {
	Source string `json:"source,omitempty"`
	URL    string `json:"url,omitempty"`
	List   string `json:"list,omitempty"`
}

// updateConfirmRequest is the body of POST /me/decks/{id}/update/confirm.
type updateConfirmRequest struct {
	PreviewID string `json:"preview_id"`
	// RequestMissing files (or extends) a deck request for the cards the
	// new version misses that the old one had not already requested.
	RequestMissing bool `json:"request_missing,omitempty"`
}

// updateDiffCard is one card row in a diff.
type updateDiffCard struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	// Sideboard marks a row in the sideboard or companion slot.
	Sideboard bool `json:"sideboard,omitempty"`
}

// updateDiffChange is a card whose copy count changed.
type updateDiffChange struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Sideboard bool   `json:"sideboard,omitempty"`
}

// updateDiff is what an update changes in the saved list.
type updateDiff struct {
	Added   []updateDiffCard   `json:"added"`
	Removed []updateDiffCard   `json:"removed"`
	Changed []updateDiffChange `json:"changed"`
	// Unchanged is how many distinct cards are the same in both.
	Unchanged        int      `json:"unchanged"`
	CommandersBefore []string `json:"commanders_before"`
	CommandersAfter  []string `json:"commanders_after"`
	CommanderChanged bool     `json:"commander_changed"`
}

// empty reports whether the lists are the same.
func (d updateDiff) empty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0 && !d.CommanderChanged
}

// updateRequestPreview is what the preview says about requesting.
type updateRequestPreview struct {
	// Available is true when this server files requests and the caller
	// can file one (a Discord identity, a configured filer).
	Available bool `json:"available"`
	// Reason says why not, for the client to show beside the box.
	Reason string `json:"reason,omitempty"`
	// NewCards are the missing cards a request would add: missing in
	// the new version and not already requested for this deck.
	NewCards []string `json:"new_cards"`
	// AlreadyRequested counts missing cards left out because the deck's
	// earlier request has them.
	AlreadyRequested int `json:"already_requested"`
}

// updatePreviewResponse is the body of POST /me/decks/{id}/update.
type updatePreviewResponse struct {
	PreviewID string    `json:"preview_id"`
	ExpiresAt time.Time `json:"expires_at"`
	// Source is "relink", "url" or "list".
	Source string `json:"source"`
	// SourceURL is the link the deck will carry after the update; empty
	// when a pasted list will clear it.
	SourceURL string `json:"source_url,omitempty"`
	// LinkCleared is true when the update drops a link the deck had.
	LinkCleared bool       `json:"link_cleared,omitempty"`
	Diff        updateDiff `json:"diff"`
	NoChange    bool       `json:"no_change"`
	// PrintingsKept counts the card rows that keep the printing they
	// had.
	PrintingsKept int `json:"printings_kept"`
	// CoverageBefore / CoverageAfter are the library headline for the
	// deck now and as it would be. Report is the new version's full
	// report: buckets, caveats and Commander violations.
	CoverageBefore *myDeckCoverageInfo  `json:"coverage_before,omitempty"`
	CoverageAfter  *myDeckCoverageInfo  `json:"coverage_after,omitempty"`
	Report         *deckcoverage.Report `json:"report"`
	Request        updateRequestPreview `json:"request"`
	// LiveTables is how many tables that have not ended are seated with
	// this deck. They keep the list they were seated with.
	LiveTables int `json:"live_tables"`
}

// Outcomes of the request an update may file (updateRequestResult.Status).
const (
	updateRequestFiled       = "filed"
	updateRequestJoined      = "joined"
	updateRequestNothingNew  = "nothing_new"
	updateRequestUnavailable = "unavailable"
	updateRequestRateLimited = "rate_limited"
	updateRequestFailed      = "failed"
)

// updateRequestResult is what the confirm says happened to the request.
type updateRequestResult struct {
	Status           string   `json:"status"`
	Message          string   `json:"message,omitempty"`
	IssueURL         string   `json:"issue_url,omitempty"`
	IssueNumber      int      `json:"issue_number,omitempty"`
	NewCards         []string `json:"new_cards,omitempty"`
	AlreadyRequested int      `json:"already_requested,omitempty"`
	RetryAfter       int      `json:"retry_after,omitempty"`
}

// updateConfirmResponse is the body of POST /me/decks/{id}/update/confirm.
type updateConfirmResponse struct {
	Deck    myDeckInfo           `json:"deck"`
	Diff    updateDiff           `json:"diff"`
	Request *updateRequestResult `json:"request,omitempty"`
}

// pendingUpdate is a previewed update waiting to be confirmed. It lives
// in memory only: nothing about it is written to the database until the
// player confirms, and a restart drops it (the confirm then asks for a
// new preview).
type pendingUpdate struct {
	id       string
	owner    uuid.UUID
	deckID   uuid.UUID
	revision string // printingRevision of the deck the preview read
	expires  time.Time

	text       string // the stored list the commit writes
	sourceURL  string // "" clears the link
	commanders []string
	cardCount  int
	diff       updateDiff
	oldKey     string
	newKey     string
	oldReport  *deckcoverage.Report // nil when the stored list would not build
	newReport  *deckcoverage.Report
	deckName   string
}

// updatePreviews holds the previews. The zero value is ready to use.
type updatePreviews struct {
	mu      sync.Mutex
	byID    map[string]*pendingUpdate
	byDeck  map[string]string // owner+deck -> preview id: one per deck
	nowFunc func() time.Time
}

func (u *updatePreviews) now() time.Time {
	if u.nowFunc != nil {
		return u.nowFunc()
	}
	return time.Now().UTC()
}

func deckSlot(owner, deckID uuid.UUID) string { return owner.String() + "/" + deckID.String() }

// put stores p under a fresh id, replacing the owner's earlier preview
// of the same deck and evicting expired previews (and, past the bound,
// the soonest to expire).
func (u *updatePreviews) put(p *pendingUpdate) error {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.byID == nil {
		u.byID, u.byDeck = map[string]*pendingUpdate{}, map[string]string{}
	}
	now := u.now()
	for id, e := range u.byID {
		if now.After(e.expires) {
			u.dropLocked(id)
		}
	}
	if old, ok := u.byDeck[deckSlot(p.owner, p.deckID)]; ok {
		u.dropLocked(old)
	}
	for len(u.byID) >= updatePreviewMax {
		oldest, first := "", true
		var at time.Time
		for id, e := range u.byID {
			if first || e.expires.Before(at) {
				oldest, at, first = id, e.expires, false
			}
		}
		u.dropLocked(oldest)
	}
	p.id = hex.EncodeToString(raw[:])
	p.expires = now.Add(updatePreviewTTL)
	u.byID[p.id] = p
	u.byDeck[deckSlot(p.owner, p.deckID)] = p.id
	return nil
}

func (u *updatePreviews) dropLocked(id string) {
	if e, ok := u.byID[id]; ok {
		delete(u.byDeck, deckSlot(e.owner, e.deckID))
		delete(u.byID, id)
	}
}

// take removes and returns the preview id names, or nil when there is
// none, it has expired, or it belongs to another person or deck. A
// preview confirms once.
func (u *updatePreviews) take(id string, owner, deckID uuid.UUID) *pendingUpdate {
	u.mu.Lock()
	defer u.mu.Unlock()
	e, ok := u.byID[id]
	if !ok || e.owner != owner || e.deckID != deckID {
		return nil
	}
	u.dropLocked(id)
	if u.now().After(e.expires) {
		return nil
	}
	return e
}

// len is how many previews are held, for tests.
func (u *updatePreviews) len() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.byID)
}

// previewUpdate handles POST /me/decks/{id}/update.
//
// The caller is a signed-in person who owns the deck, else 403 / 404
// (an id reveals nothing). The new list is validated as the first
// import validates it: a link goes through the same host allowlist and
// canonical fetch (deck.ParseDeckURL, the configured fetcher, so there
// is no way to point the server at another host), a pasted list through
// the same parser and the same size bound. Commander-rule violations are
// listed in the report and do not block the update, as they do not
// block a save (seating re-validates).
//
// Every link spends a token from the checker's per-IP fetch bucket,
// cache or not: the point of an update is to see what changed since the
// last pull, so the cache is skipped.
func (d *deckCheck) previewUpdate(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	var body updateDeckRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.Source, body.URL, body.List = strings.TrimSpace(body.Source), strings.TrimSpace(body.URL), strings.TrimSpace(body.List)
	inputs := 0
	for _, s := range []string{body.Source, body.URL, body.List} {
		if s != "" {
			inputs++
		}
	}
	switch {
	case inputs == 0:
		return httpError(http.StatusBadRequest,
			`send source "relink" to re-pull from the deck's link, a new link as url, or a pasted list as list`)
	case inputs > 1:
		return httpError(http.StatusBadRequest, "send one of source, url or list")
	case body.Source != "" && body.Source != updateRelink:
		return httpError(http.StatusBadRequest, `source must be "relink"`)
	case len(body.URL) > maxUpdateURLLen:
		return httpError(http.StatusBadRequest, "that link is too long")
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}
	if c.DeckLibrary == nil {
		return httpError(http.StatusServiceUnavailable, "deck library not configured")
	}
	saved, err := ownedLibraryDeckByID(r.Context(), c, p.UserID, r.PathValue("id"))
	if err != nil {
		return err
	}
	oldEntries, err := libraryDeckEntries(saved)
	if err != nil {
		return httpError(http.StatusUnprocessableEntity, "that saved deck could not be read: "+err.Error())
	}
	verdicts, err := deckcoverage.Verdicts(c.Cards)
	if err != nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}

	var (
		newEntries []deck.Entry
		kind       string
		sourceURL  string
		newKey     string
		source     = "text"
	)
	if body.List == "" {
		rawURL := body.URL
		kind = "url"
		if body.Source != "" {
			kind, rawURL = updateRelink, saved.SourceURL
			if rawURL == "" {
				return httpError(http.StatusBadRequest, "this deck has no link to re-pull from; send a new link or a pasted list")
			}
		}
		spend := func() bool { return d.fetchLimit == nil || d.fetchLimit.AllowRequest(r) }
		checked, err := d.deckForURLOpts(r.Context(), c, rawURL, spend, true)
		if errors.Is(err, errFetchLimited) {
			w.Header().Set("Retry-After", strconv.Itoa(fetchLimitRetryAfter))
			return writeJSON(w, http.StatusTooManyRequests, map[string]string{
				"error": "Too many deck checks from here. Wait a few seconds and try again.",
			})
		}
		if err != nil {
			return writeDeckFetchError(w, err)
		}
		// deckForURLOpts parsed the link without error.
		ref, _ := deck.ParseDeckURL(rawURL)
		newEntries, sourceURL, newKey, source = checked.entries, ref.URL(), ref.Key(), ref.Source
	} else {
		kind = "list"
		list, err := parseList(c, body.List)
		if err != nil {
			return err
		}
		newEntries, newKey = list.entries, list.key
	}
	if _, count := libraryShape(c, newEntries); count == 0 {
		return httpError(http.StatusUnprocessableEntity, "that deck has no cards")
	}

	merged, kept := carryPrintings(c, oldEntries, newEntries)
	commanders, cardCount := libraryShape(c, merged)
	oldKey, err := libraryDeckKey(c, saved, oldEntries)
	if err != nil {
		return err
	}
	oldReport, rerr := libraryDeckReport(c, saved, verdicts)
	if rerr != nil {
		oldReport = nil
	}
	newReport, err := deckcoverage.BuildWith(c.Cards, deckcoverage.Deck{
		Name: saved.Name, Source: source, SourceURL: sourceURL, DeckKey: newKey, Entries: merged,
	}, verdicts)
	if err != nil {
		return err
	}
	diff := diffEntries(c, oldEntries, merged)

	pend := &pendingUpdate{
		owner: p.UserID, deckID: saved.ID, revision: printingRevision(saved),
		text: entriesTextKeepingPrintings(c, merged), sourceURL: sourceURL,
		commanders: commanders, cardCount: cardCount, diff: diff,
		oldKey: oldKey, newKey: newKey, oldReport: oldReport, newReport: newReport, deckName: saved.Name,
	}
	if err := d.updates.put(pend); err != nil {
		c.logger().Error("deck update: could not mint a preview id", "err", err)
		return httpError(http.StatusInternalServerError, "could not prepare the update; try again")
	}

	resp := updatePreviewResponse{
		PreviewID: pend.id, ExpiresAt: pend.expires, Source: kind, SourceURL: sourceURL,
		LinkCleared: sourceURL == "" && saved.SourceURL != "",
		Diff:        diff, NoChange: diff.empty() && saved.SourceURL == sourceURL,
		PrintingsKept: kept, CoverageAfter: coverageInfoOf(newReport), Report: newReport,
		Request: d.previewRequest(r.Context(), c, p, pend),
	}
	if oldReport != nil {
		resp.CoverageBefore = coverageInfoOf(oldReport)
	}
	if c.Lobby != nil {
		resp.LiveTables = c.Lobby.TablesUsingDeck(saved.ID.String())
	}
	return writeJSON(w, http.StatusOK, resp)
}

// confirmUpdate handles POST /me/decks/{id}/update/confirm.
func (d *deckCheck) confirmUpdate(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	var body updateConfirmRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.PreviewID = strings.TrimSpace(body.PreviewID)
	if body.PreviewID == "" {
		return httpError(http.StatusBadRequest, "send the preview_id the preview returned")
	}
	if c.DeckLibrary == nil {
		return httpError(http.StatusServiceUnavailable, "deck library not configured")
	}
	saved, err := ownedLibraryDeckByID(r.Context(), c, p.UserID, r.PathValue("id"))
	if err != nil {
		return err
	}
	pend := d.updates.take(body.PreviewID, p.UserID, saved.ID)
	if pend == nil {
		return httpError(http.StatusGone, "that preview expired or was already used; preview the update again")
	}
	if pend.revision != printingRevision(saved) {
		return httpError(http.StatusConflict, "This deck changed after the preview. Preview the update again.")
	}
	updated, err := c.libraryStore().ReplaceContents(r.Context(), saved, "text", pend.text, pend.sourceURL, pend.commanders, pend.cardCount)
	switch {
	case errors.Is(err, decklibrary.ErrSourceChanged):
		return httpError(http.StatusConflict, "This deck changed after the preview. Preview the update again.")
	case errors.Is(err, decklibrary.ErrNotFound):
		return httpError(http.StatusNotFound, "deck not found")
	case err != nil:
		c.logger().Error("decklibrary: updating a saved deck failed", "err", err, "user_id", p.UserID)
		return httpError(http.StatusInternalServerError, "could not update your deck; try again")
	}

	resp := updateConfirmResponse{
		Deck: libraryDeckInfo(c, updated, libraryVerdicts(c)),
		Diff: pend.diff,
	}
	if body.RequestMissing {
		resp.Request = d.fileUpdateRequest(r.Context(), c, p, pend)
	}
	return writeJSON(w, http.StatusOK, resp)
}

// libraryDeckKey is the key a saved deck files its requests under: its
// link's deck key, else the key of its list (savedDeckForRequest's rule).
func libraryDeckKey(c Config, d decklibrary.Deck, entries []deck.Entry) (string, error) {
	if d.SourceURL != "" {
		if ref, err := deck.ParseDeckURL(d.SourceURL); err == nil {
			return ref.Key(), nil
		}
	}
	list, err := keyList(c, entries)
	if err != nil {
		return "", err
	}
	return list.key, nil
}

// --- the request an update may file ---

// missingKey identifies a card across two reports: its oracle id, else
// its lower-cased name.
func missingKey(c deckcoverage.Card) string {
	if c.OracleID != "" {
		return c.OracleID
	}
	return "name:" + strings.ToLower(c.Name)
}

// missingCards is a report's cards a request would list: the manual and
// the unreviewed buckets, in report order.
func missingCards(r *deckcoverage.Report) []deckcoverage.Card {
	if r == nil {
		return nil
	}
	return append(r.CardsIn(deckcoverage.Manual), r.CardsIn(deckcoverage.Unreviewed)...)
}

// updateDelta decides which of the new version's missing cards still
// need asking for (#2036). A card is already requested when:
//
//   - the OLD version's deck key has a request row and the old version
//     was missing that card (the old key's issue lists it), or
//   - the new version is a pasted list that has a request row of its own
//     (a list key is its content, so that issue lists this very list).
//
// Everything else the new version misses is new. It returns the new
// cards and how many were left out as already requested. The rule reads
// only the request rows ADR 0095 §3 already keeps; it is exact for one
// update after a request, and for a deck whose request was filed from
// its latest version. See the amendment for the case it does not cover.
func (d *deckCheck) updateDelta(ctx context.Context, c Config, pend *pendingUpdate) (delta []deckcoverage.Card, already int, err error) {
	missing := missingCards(pend.newReport)
	requested := map[string]bool{}
	if c.DeckRequests != nil {
		if _, lerr := c.DeckRequests.Lookup(ctx, pend.oldKey); lerr == nil {
			for _, card := range missingCards(pend.oldReport) {
				requested[missingKey(card)] = true
			}
		} else if !errors.Is(lerr, deckrequests.ErrNotFound) {
			return nil, 0, lerr
		}
		if deckcoverage.IsListKey(pend.newKey) {
			if _, lerr := c.DeckRequests.Lookup(ctx, pend.newKey); lerr == nil {
				for _, card := range missing {
					requested[missingKey(card)] = true
				}
			} else if !errors.Is(lerr, deckrequests.ErrNotFound) {
				return nil, 0, lerr
			}
		}
	}
	for _, card := range missing {
		if requested[missingKey(card)] {
			already++
			continue
		}
		delta = append(delta, card)
	}
	return delta, already, nil
}

// previewRequest fills the preview's request block.
func (d *deckCheck) previewRequest(ctx context.Context, c Config, p auth.Principal, pend *pendingUpdate) updateRequestPreview {
	out := updateRequestPreview{NewCards: []string{}}
	if c.DeckRequestFiler == nil || c.DeckRequests == nil {
		out.Reason = "Deck requests are not set up on this server."
	} else if _, err := deckRequesterFor(ctx, c, p, nil); err != nil {
		var se *httpStatusError
		if errors.As(err, &se) {
			out.Reason = se.msg
		} else {
			out.Reason = "Could not check your account for deck requests."
		}
	} else {
		out.Available = true
	}
	delta, already, err := d.updateDelta(ctx, c, pend)
	if err != nil {
		c.logger().Error("deck update: reading earlier requests failed", "err", err)
		out.Available, out.Reason = false, "Could not check this deck's earlier requests."
		delta = missingCards(pend.newReport)
	}
	for _, card := range delta {
		out.NewCards = append(out.NewCards, card.Name)
	}
	out.AlreadyRequested = already
	return out
}

// fileUpdateRequest files or extends the deck request for the cards an
// update added to what is missing. It runs after the commit and its
// failure never undoes the update: the answer says what happened.
//
// It counts as one ask against the requester's 24-hour limit, like any
// request. The deck's tracking issue is the open one under the new deck
// key or, failing that, the old one: a comment lists the new cards, and
// the new key is pointed at the same issue so later updates and checks
// join it. With no open issue, a new one is filed whose checklist holds
// only the new cards.
func (d *deckCheck) fileUpdateRequest(ctx context.Context, c Config, p auth.Principal, pend *pendingUpdate) *updateRequestResult {
	if c.DeckRequestFiler == nil || c.DeckRequests == nil {
		return &updateRequestResult{Status: updateRequestUnavailable, Message: "Deck requests are not set up on this server."}
	}
	who, err := deckRequesterFor(ctx, c, p, nil)
	if err != nil {
		msg := "Could not check your account for deck requests."
		var se *httpStatusError
		if errors.As(err, &se) {
			msg = se.msg
		}
		return &updateRequestResult{Status: updateRequestUnavailable, Message: msg}
	}
	exempt := deckRequestExempt(ctx, c, p, false, who)
	requesterKey := "discord:" + who.DiscordID

	d.fileMu.Lock()
	defer d.fileMu.Unlock()
	delta, already, err := d.updateDelta(ctx, c, pend)
	if err != nil {
		c.logger().Error("deck update: reading earlier requests failed", "err", err)
		return &updateRequestResult{Status: updateRequestFailed, Message: "Could not check this deck's earlier requests; request from the deck's card instead."}
	}
	names := make([]string, 0, len(delta))
	for _, card := range delta {
		names = append(names, card.Name)
	}
	if len(delta) == 0 {
		return &updateRequestResult{Status: updateRequestNothingNew, AlreadyRequested: already}
	}
	if !exempt {
		limited, secs, err := d.askLimit(ctx, c, requesterKey)
		if err != nil {
			c.logger().Error("deck update: reading the rate limit failed", "err", err)
			return &updateRequestResult{Status: updateRequestFailed, Message: "Could not check your request limit; request from the deck's card instead."}
		}
		if limited {
			return &updateRequestResult{Status: updateRequestRateLimited, RetryAfter: secs, NewCards: names, AlreadyRequested: already}
		}
	}

	keys := []string{pend.newKey}
	if pend.oldKey != pend.newKey {
		keys = append(keys, pend.oldKey)
	}
	for _, key := range keys {
		existing, lerr := c.DeckRequests.Lookup(ctx, key)
		if errors.Is(lerr, deckrequests.ErrNotFound) {
			continue
		}
		if lerr != nil {
			c.logger().Error("deck update: request lookup failed", "deck", key, "err", lerr)
			return &updateRequestResult{Status: updateRequestFailed, Message: "Could not check this deck's earlier requests; request from the deck's card instead."}
		}
		open, url, ierr := d.issueOpen(ctx, c, existing)
		if ierr != nil {
			return &updateRequestResult{Status: updateRequestFailed, Message: fmt.Sprintf("Checking the deck's existing issue failed: %s", ierr)}
		}
		if !open {
			continue
		}
		if _, cerr := c.DeckRequestFiler.CreateComment(ctx, existing.IssueNumber, renderDeckUpdateComment(who, pend.newReport, delta, already)); cerr != nil {
			return &updateRequestResult{Status: updateRequestFailed, Message: fmt.Sprintf("Commenting on the deck's issue failed: %s", cerr)}
		}
		if key != pend.newKey {
			// The updated deck now answers to the new key; point it at the
			// issue that tracks it, keeping the original ask window.
			if perr := c.DeckRequests.Put(ctx, deckrequests.Request{
				DeckKey: pend.newKey, IssueNumber: existing.IssueNumber, IssueURL: existing.IssueURL, CreatedAt: existing.CreatedAt,
			}); perr != nil {
				c.logger().Error("deck update: pointing the new key at the issue failed", "deck", pend.newKey, "err", perr)
			}
		}
		d.recordAsk(ctx, c, pend.newKey, requesterKey)
		return &updateRequestResult{Status: updateRequestJoined, IssueURL: url, IssueNumber: existing.IssueNumber, NewCards: names, AlreadyRequested: already}
	}

	only := map[string]bool{}
	for _, card := range delta {
		only[missingKey(card)] = true
	}
	url, number, ferr := fileDeckRequestIssue(ctx, c, deckRequestTitle(pend.newReport), renderDeckRequestIssueOnly(who, false, pend.newReport, only))
	if ferr != nil {
		return &updateRequestResult{Status: updateRequestFailed, Message: fmt.Sprintf("Filing the issue failed: %s", ferr)}
	}
	if perr := c.DeckRequests.Put(ctx, deckrequests.Request{
		DeckKey: pend.newKey, IssueNumber: number, IssueURL: url, CreatedAt: d.now(),
	}); perr != nil {
		c.logger().Error("deck update: issue filed but its row was not saved", "deck", pend.newKey, "issue", url, "err", perr)
	}
	d.recordAsk(ctx, c, pend.newKey, requesterKey)
	return &updateRequestResult{Status: updateRequestFiled, IssueURL: url, IssueNumber: number, NewCards: names, AlreadyRequested: already}
}

// renderDeckUpdateComment is the comment an update adds to a deck's open
// request: who updated it, and the checklist of cards that are new to
// the request. Names come off the card index; the display name goes
// through mdUserText.
func renderDeckUpdateComment(who deckRequester, r *deckcoverage.Report, delta []deckcoverage.Card, already int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "The deck was updated by %s. %d more card%s to add:\n\n",
		mdUserText(who.DisplayName, deckRequesterNameMax), len(delta), plural(len(delta)))
	for _, c := range delta {
		fmt.Fprintf(&b, "- [ ] %s (%s)\n", c.Name, c.OracleID)
	}
	if already > 0 {
		fmt.Fprintf(&b, "\n%d other missing card%s already requested here %s not repeated.\n",
			already, plural(already), map[bool]string{true: "is", false: "are"}[already == 1])
	}
	var parts []string
	for _, bucket := range deckcoverage.Buckets {
		parts = append(parts, fmt.Sprintf("%s: %d", strings.ToLower(deckBucketLabels[bucket]), r.Counts[bucket]))
	}
	fmt.Fprintf(&b, "\nCurrent counts — %s.\n", strings.Join(parts, "; "))
	return b.String()
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// --- the diff and the stored text ---

type diffKey struct {
	name string
	side bool
}

// tally sums copies per (card, board) and remembers a display name.
func tally(c Config, entries []deck.Entry) (map[diffKey]int, map[diffKey]string) {
	counts, names := map[diffKey]int{}, map[diffKey]string{}
	for _, e := range entries {
		if e.Count <= 0 {
			continue
		}
		name := strings.TrimSpace(e.Name)
		if c.Cards != nil {
			if card, ok := c.Cards.FindByName(name); ok {
				name = card.Name
			}
		}
		k := diffKey{name: strings.ToLower(name), side: e.IsSideboard || e.IsCompanion}
		counts[k] += e.Count
		if _, ok := names[k]; !ok {
			names[k] = name
		}
	}
	return counts, names
}

// diffEntries compares two lists by card, then by board (the deck, or
// the sideboard and companion slot). A card whose copies changed is one
// Changed row; a basic land going from ten to twelve is not an add.
func diffEntries(c Config, before, after []deck.Entry) updateDiff {
	bc, bn := tally(c, before)
	ac, an := tally(c, after)
	d := updateDiff{Added: []updateDiffCard{}, Removed: []updateDiffCard{}, Changed: []updateDiffChange{}}
	for k, n := range ac {
		switch had := bc[k]; {
		case had == 0:
			d.Added = append(d.Added, updateDiffCard{Name: an[k], Count: n, Sideboard: k.side})
		case had != n:
			d.Changed = append(d.Changed, updateDiffChange{Name: an[k], From: had, To: n, Sideboard: k.side})
		default:
			d.Unchanged++
		}
	}
	for k, n := range bc {
		if ac[k] == 0 {
			d.Removed = append(d.Removed, updateDiffCard{Name: bn[k], Count: n, Sideboard: k.side})
		}
	}
	sort.Slice(d.Added, func(i, j int) bool {
		return lessCard(d.Added[i].Name, d.Added[i].Sideboard, d.Added[j].Name, d.Added[j].Sideboard)
	})
	sort.Slice(d.Removed, func(i, j int) bool {
		return lessCard(d.Removed[i].Name, d.Removed[i].Sideboard, d.Removed[j].Name, d.Removed[j].Sideboard)
	})
	sort.Slice(d.Changed, func(i, j int) bool {
		return lessCard(d.Changed[i].Name, d.Changed[i].Sideboard, d.Changed[j].Name, d.Changed[j].Sideboard)
	})
	d.CommandersBefore, _ = libraryShape(c, before)
	d.CommandersAfter, _ = libraryShape(c, after)
	d.CommanderChanged = !sameStrings(d.CommandersBefore, d.CommandersAfter)
	return d
}

func lessCard(an string, aside bool, bn string, bside bool) bool {
	if aside != bside {
		return !aside
	}
	return strings.ToLower(an) < strings.ToLower(bn)
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// carryPrintings gives each new row the printing its card had in the
// old list, when it had one chosen: an unchanged card keeps its art
// across an update (#2036). Rows match on board and card name, and the
// n-th new row of a card takes the n-th old row's printing, so a deck
// with distinct basic-land printings keeps them. A card the old list
// did not print, or a new row beyond the old rows, keeps whatever the
// new source says. kept counts the rows that took an old printing.
func carryPrintings(c Config, old, fresh []deck.Entry) ([]deck.Entry, int) {
	rowKey := func(e deck.Entry) string {
		name := strings.TrimSpace(e.Name)
		if c.Cards != nil {
			if card, ok := c.Cards.FindByName(name); ok {
				name = card.Name
			}
		}
		return printingBoard(e) + "\x00" + strings.ToLower(name)
	}
	// A row without a printing in the old list still occupies its place
	// in the order, or the n-th new row would drift onto the wrong old
	// one, so every old row is indexed, not only those with a printing.
	seenOld := map[string]int{}
	oldRows := map[string][]deck.Entry{}
	for _, e := range old {
		if e.Count > 0 {
			oldRows[rowKey(e)] = append(oldRows[rowKey(e)], e)
		}
	}
	out := make([]deck.Entry, len(fresh))
	copy(out, fresh)
	kept := 0
	for i := range out {
		if out[i].Count <= 0 {
			continue
		}
		k := rowKey(out[i])
		n := seenOld[k]
		seenOld[k]++
		if n >= len(oldRows[k]) {
			continue
		}
		o := oldRows[k][n]
		if o.PrintingID == uuid.Nil && (o.SetCode == "" || o.CollectorNo == "") {
			continue
		}
		out[i].PrintingID, out[i].SetCode, out[i].CollectorNo = o.PrintingID, o.SetCode, o.CollectorNo
		kept++
	}
	return out, kept
}

// entriesTextKeepingPrintings renders entries as the stored list: one
// section per board in a fixed order, each line "N Name" with the
// printing when the entry has one (the same line the printing picker
// writes). ParseText reads it back to the same entries.
func entriesTextKeepingPrintings(c Config, entries []deck.Entry) string {
	var b strings.Builder
	for _, board := range []string{"Commander", "Companion", "Mainboard", "Sideboard"} {
		wrote := false
		for _, e := range entries {
			if e.Count <= 0 || printingBoard(e) != board {
				continue
			}
			if !wrote {
				b.WriteString(board + ":\n")
				wrote = true
			}
			b.WriteString(entryLine(c, e) + "\n")
		}
	}
	return b.String()
}
