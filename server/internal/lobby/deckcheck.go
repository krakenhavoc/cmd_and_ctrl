package lobby

// deckcheck.go serves ADR 0095's two deck routes:
//
//   - POST /deck-coverage — public, no session. A Moxfield or
//     Archidekt link (or a pasted list) in, a deckcoverage.Report out:
//     how much of the deck the engine automates.
//   - POST /deck-requests — signed in (a user with a Discord identity,
//     or the bot's admin session naming a Discord member). Files the
//     deck's missing cards as a GitHub issue, or joins the open issue
//     already filed for that deck. A link or, since ADR 0095's
//     2026-09-25 amendment, a pasted list: Moxfield blocks this
//     server, so a Moxfield player's only route is Export and paste.
//
// The report is deckcoverage.Build's and nothing here re-decides a
// bucket. What this file owns is the HTTP edge: the per-IP limit on a
// route that fetches third-party URLs for anonymous callers, the
// ten-minute report cache, and ADR 0095 §3's filing rules.

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deckrequests"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/github"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/redact"
)

// DeckRequestFiler is the GitHub half of POST /deck-requests.
// Satisfied by *github.Client; narrow so handler tests can record
// what would have been filed without talking to GitHub (the
// BugReporter seam's pattern). Nil in Config turns the route off: it
// answers 503 naming CMDCTRL_GITHUB_TOKEN.
type DeckRequestFiler interface {
	CreateIssue(ctx context.Context, title, body string, labels []string) (url string, number int, err error)
	GetIssue(ctx context.Context, number int) (github.Issue, error)
	CreateComment(ctx context.Context, number int, body string) (url string, err error)
}

// DeckFetcher fetches a deck by URL. deck.FetchFromURL over
// Config.DeckHTTPClient in production; tests inject a fake so nothing
// reaches the network.
type DeckFetcher func(ctx context.Context, rawURL string) (name string, entries []deck.Entry, err error)

const (
	// deckReportTTL is how long a fetched deck's report is reused
	// (ADR 0095 §2): two checks of one deck fetch it once.
	deckReportTTL = 10 * time.Minute
	// deckReportCacheMax bounds the cache. A report is a few KiB.
	deckReportCacheMax = 256

	// deckRequestWindow / deckRequestLimit are ADR 0095 §3 step 1:
	// three asks per requester per rolling 24 hours, counted from the
	// database so a deploy does not reset them.
	deckRequestWindow = 24 * time.Hour
	deckRequestLimit  = 3

	// deckRequestTitleMax bounds the issue title; deckRequesterNameMax
	// bounds a display name the bot passes in.
	deckRequestTitleMax  = 200
	deckRequesterNameMax = 80

	deckRequestTitlePrefix = "[deck-request] "
)

// deckRequestLabels are the labels a deck-request issue is filed
// with. A label GitHub refuses costs the labels, never the issue.
var deckRequestLabels = []string{"enhancement", "deck-request"}

// deckCheck is the state the two routes share across requests: the
// report cache and the lock that keeps two asks for one deck from
// filing two issues. Built once per Handler.
type deckCheck struct {
	cache *deckReportCache
	// fileMu serialises the lookup-then-file-or-comment step of a
	// deck request. One server process, a handful of users: a single
	// lock is the whole concurrency story.
	fileMu sync.Mutex
	now    func() time.Time
}

func newDeckCheck() *deckCheck {
	now := func() time.Time { return time.Now().UTC() }
	return &deckCheck{cache: newDeckReportCache(deckReportTTL, deckReportCacheMax, now), now: now}
}

// --- POST /deck-coverage ---

type deckCoverageRequest struct {
	URL  string `json:"url,omitempty"`
	Text string `json:"text,omitempty"`
}

// deckCoverageLimit applies ADR 0095 §2's per-IP limit to the public
// checker — except for an admin session, which is the Discord bot
// calling over loopback on behalf of every guild member at once and
// would otherwise share one tiny bucket with all of them. Admin calls
// ride their own bucket instead. Any other credential, valid or not,
// is ignored: the route is public, and a session buys nothing.
func deckCoverageLimit(c Config, public, admin *ratelimit.Limiter, next http.Handler) http.Handler {
	publicNext := public.Middleware(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cred := auth.CredentialFromRequest(r); cred != "" && c.Auth != nil {
			if p, err := c.Auth.Validate(r.Context(), cred); err == nil && p.Role == auth.RoleAdmin {
				if !admin.Allow("admin") {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
					_, _ = w.Write([]byte(`{"error":"too many requests"}`))
					return
				}
				next.ServeHTTP(w, r)
				return
			}
		}
		publicNext.ServeHTTP(w, r)
	})
}

func (d *deckCheck) coverage(c Config, w http.ResponseWriter, r *http.Request) error {
	var body deckCoverageRequest
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
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}

	var (
		report *deckcoverage.Report
		err    error
	)
	if body.URL != "" {
		report, err = d.reportForURL(r.Context(), c, body.URL)
	} else {
		report, err = d.reportForText(c, body.Text)
	}
	if err != nil {
		return writeDeckFetchError(w, err)
	}
	return writeJSON(w, http.StatusOK, report)
}

// parsedList is a pasted list read and keyed: its entries and its
// "list:<hash>" deck key (deckcoverage.ListKey).
type parsedList struct {
	key     string
	entries []deck.Entry
}

// parseList reads a pasted list and derives its deck key. The key is
// the list's identity (ADR 0095, amendment 2026-09-25): the same deck
// pasted from two exports is one key, so it is one cache entry and one
// deck request.
func parseList(c Config, text string) (parsedList, error) {
	entries, err := deck.ParseText(text)
	if err != nil {
		return parsedList{}, httpError(http.StatusBadRequest, "could not read that list: "+err.Error())
	}
	key, err := deckcoverage.ListKey(c.Cards, entries)
	switch {
	case errors.Is(err, deckcoverage.ErrNoIndex):
		return parsedList{}, httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	case errors.Is(err, deckcoverage.ErrListTooLong):
		return parsedList{}, httpError(http.StatusBadRequest,
			fmt.Sprintf("that list names more than %d cards; paste one deck at a time", deckcoverage.MaxListCopies))
	case err != nil:
		return parsedList{}, err
	}
	return parsedList{key: key, entries: entries}, nil
}

// reportForText builds the report for a pasted list, reusing a cached
// one for the same list key. A pasted list has no name and no link:
// its report says Source "text" and carries the list key.
func (d *deckCheck) reportForText(c Config, text string) (*deckcoverage.Report, error) {
	list, err := parseList(c, text)
	if err != nil {
		return nil, err
	}
	return d.reportForList(c, list)
}

func (d *deckCheck) reportForList(c Config, list parsedList) (*deckcoverage.Report, error) {
	if r, ok := d.cache.get(list.key); ok {
		return r, nil
	}
	r, err := buildReport(c, deckcoverage.Deck{Source: "text", DeckKey: list.key, Entries: list.entries})
	if err != nil {
		return nil, err
	}
	d.cache.put(list.key, r)
	return r, nil
}

// reportForURL returns the report for the deck a link names: from the
// cache when it is fresh, otherwise fetched and built and then cached.
// The fetch always goes to the deck's canonical URL, so the fetcher
// never sees a caller's query string.
func (d *deckCheck) reportForURL(ctx context.Context, c Config, rawURL string) (*deckcoverage.Report, error) {
	ref, err := deck.ParseDeckURL(rawURL)
	if err != nil {
		return nil, newDeckFetchError(rawURL, err)
	}
	if r, ok := d.cache.get(ref.Key()); ok {
		return r, nil
	}
	name, entries, err := c.deckFetcher()(ctx, ref.URL())
	if err != nil {
		return nil, newDeckFetchError(rawURL, err)
	}
	r, err := buildReport(c, deckcoverage.Deck{
		Name:      name,
		Source:    ref.Source,
		SourceURL: ref.URL(),
		DeckKey:   ref.Key(),
		Entries:   entries,
	})
	if err != nil {
		return nil, err
	}
	d.cache.put(ref.Key(), r)
	return r, nil
}

func buildReport(c Config, d deckcoverage.Deck) (*deckcoverage.Report, error) {
	r, err := deckcoverage.Build(c.Cards, d)
	if errors.Is(err, deckcoverage.ErrNoIndex) {
		return nil, httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}
	return r, err
}

// deckFetcher returns the configured fetcher, or deck.FetchFromURL.
func (c Config) deckFetcher() DeckFetcher {
	if c.FetchDeck != nil {
		return c.FetchDeck
	}
	client := c.DeckHTTPClient
	if client == nil {
		client = deck.DefaultClient()
	}
	return func(ctx context.Context, rawURL string) (string, []deck.Entry, error) {
		return deck.FetchFromURL(ctx, client, rawURL)
	}
}

// deckFetchError is a deck link that could not be read, with a status
// and a sentence a player can act on. The typed violation rides along
// so a client can key on its code.
type deckFetchError struct {
	status    int
	msg       string
	violation deck.Violation
	// hint is deckFetchHintPasteList when the way forward is to paste
	// the list instead of linking it. Empty otherwise.
	hint string
}

// deckFetchHintPasteList is the `hint` a Moxfield fetch error carries
// (ADR 0095, amendment 2026-09-25): Moxfield blocks this server, so
// the only way to check or request a Moxfield deck is to paste it. A
// client keys its "paste the list instead" affordance on it.
const deckFetchHintPasteList = "paste_list"

// moxfieldBlockedMessage is the sentence every Moxfield fetch error
// shows. Moxfield answers this server with a Cloudflare 403 on every
// endpoint (the block keys on the client's fingerprint, not its IP),
// so a "not found" or "private" from it is not worth believing either.
const moxfieldBlockedMessage = "Moxfield blocks our server. On Moxfield, open the deck → Export → Copy plain text, then paste the list here instead."

func (e *deckFetchError) Error() string { return e.msg }

func newDeckFetchError(rawURL string, err error) error {
	v, ok := deck.FetchViolation(rawURL, err)
	if !ok {
		// A payload the parser refused (an empty Archidekt export, a
		// Moxfield companion slot): the deck was reached but is not a
		// list we can read.
		return withMoxfieldHint(rawURL, &deckFetchError{
			status:    http.StatusUnprocessableEntity,
			msg:       "Could not read that deck: " + err.Error(),
			violation: deck.Violation{Code: "unreadable_deck", Card: rawURL, Message: err.Error()},
		})
	}
	e := &deckFetchError{violation: v}
	switch v.Code {
	case deck.CodeUnknownSource:
		e.status = http.StatusBadRequest
		e.msg = "That link is not a Moxfield or Archidekt deck. Use a link like https://moxfield.com/decks/… or https://archidekt.com/decks/…, or paste the list as text."
	case deck.CodeDeckNotFound:
		e.status = http.StatusNotFound
		e.msg = "That deck was not found. It may have been deleted, or the link has a typo."
	case deck.CodeDeckPrivate:
		e.status = http.StatusUnprocessableEntity
		e.msg = "That deck is private. Make it public or unlisted on the deck site, or paste the list as text."
	case deck.CodeUpstreamBlocked:
		e.status = http.StatusBadGateway
		e.msg = "The deck site is blocking this server right now. Try an Archidekt link, or paste the list as text."
	default: // deck.CodeExternalAPIUnavailable
		e.status = http.StatusBadGateway
		e.msg = "The deck site did not answer. Try again in a minute, or paste the list as text."
	}
	return withMoxfieldHint(rawURL, e)
}

// withMoxfieldHint rewrites a fetch error for a Moxfield link to
// moxfieldBlockedMessage with the paste_list hint, keeping its status
// and code. A link that is not a Moxfield deck is left alone.
func withMoxfieldHint(rawURL string, e *deckFetchError) *deckFetchError {
	if e.violation.Code == deck.CodeUnknownSource {
		return e
	}
	if ref, err := deck.ParseDeckURL(rawURL); err != nil || ref.Source != deck.SourceMoxfield {
		return e
	}
	e.msg = moxfieldBlockedMessage
	e.hint = deckFetchHintPasteList
	return e
}

// writeDeckFetchError writes a deckFetchError as
// {error, code, violations: [v]} (plus `hint` when it has one), and
// hands anything else back to the ordinary lobby error path.
func writeDeckFetchError(w http.ResponseWriter, err error) error {
	var fe *deckFetchError
	if !errors.As(err, &fe) {
		return err
	}
	body := map[string]any{
		"error":      fe.msg,
		"code":       fe.violation.Code,
		"violations": []deck.Violation{fe.violation},
	}
	if fe.hint != "" {
		body["hint"] = fe.hint
	}
	return writeJSON(w, fe.status, body)
}

// --- POST /deck-requests ---

type deckRequestBody struct {
	// URL is a Moxfield or Archidekt link; Text is a pasted list.
	// Exactly one of them (ADR 0095, amendment 2026-09-25).
	URL       string         `json:"url,omitempty"`
	Text      string         `json:"text,omitempty"`
	Requester *deckRequester `json:"requester,omitempty"`
}

// deckRequester is who asked. The bot names them; a signed-in user's
// comes from their session and the users table.
type deckRequester struct {
	DiscordID   string `json:"discord_id"`
	DisplayName string `json:"display_name"`
}

// Deck request outcomes (the response's status field).
const (
	deckRequestFiled        = "filed"
	deckRequestJoined       = "joined"
	deckRequestNothingToAdd = "nothing_to_add"
	deckRequestRateLimited  = "rate_limited"
)

type deckRequestResponse struct {
	Status      string `json:"status"`
	IssueURL    string `json:"issue_url,omitempty"`
	IssueNumber int    `json:"issue_number,omitempty"`
	// RetryAfter is whole seconds until the requester may ask again.
	// rate_limited only.
	RetryAfter int `json:"retry_after,omitempty"`
	// AlreadyRequested marks a joined answer for someone who had
	// already asked on this issue: no comment was added.
	AlreadyRequested bool `json:"already_requested,omitempty"`
	// Report is the coverage the request was decided on. Absent only
	// for rate_limited, which is decided before the deck is fetched.
	Report *deckcoverage.Report `json:"report,omitempty"`
}

func (d *deckCheck) request(c Config, w http.ResponseWriter, r *http.Request) error {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if c.DeckRequestFiler == nil {
		return httpError(http.StatusServiceUnavailable,
			"deck requests are not configured on this server (CMDCTRL_GITHUB_TOKEN is not set)")
	}
	if c.DeckRequests == nil {
		return httpError(http.StatusServiceUnavailable,
			"deck requests are not configured on this server (they need its database; CMDCTRL_DATA_DIR is not set)")
	}

	var body deckRequestBody
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	body.URL, body.Text = strings.TrimSpace(body.URL), strings.TrimSpace(body.Text)
	switch {
	case body.URL == "" && body.Text == "":
		return httpError(http.StatusBadRequest,
			"send the deck's Moxfield or Archidekt link as url, or a pasted list as text")
	case body.URL != "" && body.Text != "":
		return httpError(http.StatusBadRequest, "send url or text, not both")
	}
	who, err := deckRequesterFor(r.Context(), c, p, body.Requester)
	if err != nil {
		return err
	}
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "the card index is not loaded on this server")
	}

	// The deck key: the link's deck ID, or the pasted list's hash. A
	// list is parsed and keyed before the limit, because that is cheap
	// and touches nothing outside this process; a link is not fetched
	// until the limit has passed.
	var (
		key  string
		list parsedList
	)
	if body.URL != "" {
		ref, err := deck.ParseDeckURL(body.URL)
		if err != nil {
			return writeDeckFetchError(w, newDeckFetchError(body.URL, err))
		}
		key = ref.Key()
	} else {
		if list, err = parseList(c, body.Text); err != nil {
			return err
		}
		key = list.key
	}
	requesterKey := "discord:" + who.DiscordID

	// Refuse an over-limit ask before fetching anything; checked again
	// under the lock, where it is authoritative.
	if limited, err := d.rateLimited(r.Context(), c, w, requesterKey); limited || err != nil {
		return err
	}

	var report *deckcoverage.Report
	if body.URL != "" {
		report, err = d.reportForURL(r.Context(), c, body.URL)
	} else {
		report, err = d.reportForList(c, list)
	}
	if err != nil {
		return writeDeckFetchError(w, err)
	}
	if !report.Needed() {
		return writeJSON(w, http.StatusOK, deckRequestResponse{Status: deckRequestNothingToAdd, Report: report})
	}

	d.fileMu.Lock()
	defer d.fileMu.Unlock()
	if limited, err := d.rateLimited(r.Context(), c, w, requesterKey); limited || err != nil {
		return err
	}

	existing, err := c.DeckRequests.Lookup(r.Context(), key)
	switch {
	case errors.Is(err, deckrequests.ErrNotFound):
		// First request for this deck: file below.
	case err != nil:
		c.logger().Error("deck request: lookup failed", "deck", key, "err", err)
		return httpError(http.StatusInternalServerError, "could not check this deck's earlier requests; try again")
	default:
		open, url, err := d.issueOpen(r.Context(), c, existing)
		if err != nil {
			return httpError(http.StatusBadGateway, fmt.Sprintf("checking the deck's existing issue failed: %s", err))
		}
		if open {
			return d.join(r.Context(), c, w, existing, url, requesterKey, who, report)
		}
		// Closed (or deleted): a new issue, and the row repointed.
	}
	return d.file(r.Context(), c, w, key, requesterKey, who, p, report)
}

// deckRequesterFor decides who is asking. An admin session is the bot
// and must name the Discord member; anyone else is asking for
// themselves and must be signed in with Discord.
func deckRequesterFor(ctx context.Context, c Config, p auth.Principal, named *deckRequester) (deckRequester, error) {
	if p.Role == auth.RoleAdmin {
		if named == nil {
			return deckRequester{}, httpError(http.StatusBadRequest,
				"an admin session must name the requester: requester {discord_id, display_name}")
		}
		id := strings.TrimSpace(named.DiscordID)
		if !isSnowflake(id) {
			return deckRequester{}, httpError(http.StatusBadRequest, "requester.discord_id must be a Discord user id")
		}
		name := cleanDisplayName(named.DisplayName)
		if name == "" {
			return deckRequester{}, httpError(http.StatusBadRequest, "requester.display_name is required")
		}
		return deckRequester{DiscordID: id, DisplayName: name}, nil
	}
	if named != nil {
		return deckRequester{}, httpError(http.StatusForbidden,
			"requester is admin-only; a signed-in request is filed under your own Discord account")
	}
	if p.UserID == uuid.Nil {
		return deckRequester{}, httpError(http.StatusForbidden, "sign in with Discord to request cards")
	}
	store := c.userStore()
	id, err := store.DiscordSubject(ctx, p.UserID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return deckRequester{}, httpError(http.StatusForbidden,
				"your account has no Discord identity; sign in with Discord to request cards")
		}
		c.logger().Error("deck request: resolving the requester's Discord id failed", "err", err)
		return deckRequester{}, httpError(http.StatusInternalServerError, "could not look you up; try again")
	}
	name := ""
	if u, err := store.Get(ctx, p.UserID); err == nil {
		name = u.DisplayName
	}
	for _, fallback := range []string{p.DiscordGlobalName, p.DiscordUsername, p.Name} {
		if name != "" {
			break
		}
		name = fallback
	}
	name = cleanDisplayName(name)
	if name == "" {
		name = "a Discord user"
	}
	return deckRequester{DiscordID: id, DisplayName: name}, nil
}

// rateLimited answers ADR 0095 §3 step 1. When the requester is over
// the limit it writes the 429 itself and reports true.
func (d *deckCheck) rateLimited(ctx context.Context, c Config, w http.ResponseWriter, requesterKey string) (bool, error) {
	now := d.now()
	asks, err := c.DeckRequests.AsksSince(ctx, requesterKey, now.Add(-deckRequestWindow))
	if err != nil {
		c.logger().Error("deck request: reading the rate limit failed", "err", err)
		return true, httpError(http.StatusInternalServerError, "could not check your request limit; try again")
	}
	if len(asks) < deckRequestLimit {
		return false, nil
	}
	// The ask that has to age out before the count drops below the
	// limit again.
	free := asks[len(asks)-deckRequestLimit].Add(deckRequestWindow)
	secs := int(math.Ceil(free.Sub(now).Seconds()))
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	return true, writeJSON(w, http.StatusTooManyRequests, deckRequestResponse{
		Status:     deckRequestRateLimited,
		RetryAfter: secs,
	})
}

// issueOpen reports whether the deck's current issue is still open. A
// deleted or transferred-away issue (404 / 410) counts as closed.
func (d *deckCheck) issueOpen(ctx context.Context, c Config, existing deckrequests.Request) (bool, string, error) {
	iss, err := c.DeckRequestFiler.GetIssue(ctx, existing.IssueNumber)
	if err != nil {
		var sc statusCoder
		if errors.As(err, &sc) && (sc.StatusCode() == http.StatusNotFound || sc.StatusCode() == http.StatusGone) {
			return false, "", nil
		}
		return false, "", err
	}
	url := existing.IssueURL
	if iss.HTMLURL != "" {
		url = iss.HTMLURL
	}
	return iss.Open(), url, nil
}

// join adds the requester to the deck's open issue: a comment with the
// current counts, unless this requester already asked on this issue.
func (d *deckCheck) join(ctx context.Context, c Config, w http.ResponseWriter, existing deckrequests.Request,
	url, requesterKey string, who deckRequester, report *deckcoverage.Report) error {
	resp := deckRequestResponse{Status: deckRequestJoined, IssueURL: url, IssueNumber: existing.IssueNumber, Report: report}
	asked, err := c.DeckRequests.HasAsked(ctx, existing.DeckKey, requesterKey, existing.CreatedAt)
	if err != nil {
		c.logger().Error("deck request: reading earlier asks failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not check this deck's earlier requests; try again")
	}
	if asked {
		resp.AlreadyRequested = true
		return writeJSON(w, http.StatusOK, resp)
	}
	if _, err := c.DeckRequestFiler.CreateComment(ctx, existing.IssueNumber, renderDeckRequestComment(who, report)); err != nil {
		return httpError(http.StatusBadGateway, fmt.Sprintf("commenting on the deck's issue failed: %s", err))
	}
	d.recordAsk(ctx, c, existing.DeckKey, requesterKey)
	return writeJSON(w, http.StatusOK, resp)
}

// file files a new issue for the deck and points its row at it.
func (d *deckCheck) file(ctx context.Context, c Config, w http.ResponseWriter, key, requesterKey string,
	who deckRequester, p auth.Principal, report *deckcoverage.Report) error {
	title := deckRequestTitle(report)
	body := renderDeckRequestIssue(who, p.Role == auth.RoleAdmin, report)
	url, number, err := fileDeckRequestIssue(ctx, c, title, body)
	if err != nil {
		return httpError(http.StatusBadGateway, fmt.Sprintf("filing the issue failed: %s", err))
	}
	if err := c.DeckRequests.Put(ctx, deckrequests.Request{
		DeckKey: key, IssueNumber: number, IssueURL: url, CreatedAt: d.now(),
	}); err != nil {
		// The issue exists; losing the row costs deduplication for this
		// deck, not the request. Log it loudly and answer success.
		c.logger().Error("deck request: issue filed but its row was not saved", "deck", key, "issue", url, "err", err)
	}
	d.recordAsk(ctx, c, key, requesterKey)
	return writeJSON(w, http.StatusCreated, deckRequestResponse{
		Status: deckRequestFiled, IssueURL: url, IssueNumber: number, Report: report,
	})
}

func (d *deckCheck) recordAsk(ctx context.Context, c Config, key, requesterKey string) {
	if err := c.DeckRequests.RecordAsk(ctx, key, requesterKey, d.now()); err != nil {
		c.logger().Error("deck request: recording the ask failed", "deck", key, "err", err)
	}
}

// fileDeckRequestIssue files with the deck-request labels, and re-files
// unlabelled when GitHub refuses them — fileBugIssue's rule (ADR 0017
// §8): a label must never cost the request.
func fileDeckRequestIssue(ctx context.Context, c Config, title, body string) (string, int, error) {
	url, number, err := c.DeckRequestFiler.CreateIssue(ctx, title, body, deckRequestLabels)
	if err == nil || !bugRequestRejected(err) {
		return url, number, err
	}
	url, number, retryErr := c.DeckRequestFiler.CreateIssue(ctx, title, body, nil)
	if retryErr != nil {
		return "", 0, err
	}
	c.logger().Error("deck request: labels rejected by GitHub — issue filed unlabelled",
		"labels", strings.Join(deckRequestLabels, ","), "issue", url, "err", err)
	return url, number, nil
}

// isSnowflake reports whether s looks like a Discord user id: 1-20
// decimal digits.
func isSnowflake(s string) bool {
	if s == "" || len(s) > 20 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func cleanDisplayName(s string) string {
	s = strings.TrimSpace(strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(s))
	if len(s) > deckRequesterNameMax {
		s = strings.TrimSpace(s[:deckRequesterNameMax])
	}
	return s
}

// --- the issue ---

// mdUserText makes a string someone else chose safe to publish in an
// issue: secrets redacted (ADR 0017's rule for anything posted to
// GitHub), one line, bounded, an @ that cannot ping anybody, and
// markdown escaped — a deck named "x](https://…)" is text in our
// public repo, never a link, an image or a table cell break.
func mdUserText(s string, n int) string {
	s = redact.Secrets(s)
	s = strings.NewReplacer("\n", " ", "\r", " ").Replace(s)
	s = clip(s, n)
	s = mdEscaper.Replace(s)
	s = strings.ReplaceAll(s, "@", "@​")
	return strings.TrimSpace(s)
}

// mdEscaper backslash-escapes the markdown characters that could turn
// published user text into a link, an image, emphasis, code, HTML or a
// table cell.
var mdEscaper = strings.NewReplacer(
	`\`, `\\`, "[", `\[`, "]", `\]`, "(", `\(`, ")", `\)`,
	"<", `\<`, ">", `\>`, "`", "\\`", "*", `\*`, "_", `\_`,
	"|", `\|`, "!", `\!`, "~", `\~`,
)

// deckRequestTitle is "[deck-request] <deck name>", falling back to
// the commander when the deck has no name — always, for a pasted list,
// which has none — and then to the deck key.
func deckRequestTitle(r *deckcoverage.Report) string {
	name := strings.TrimSpace(r.DeckName)
	if name == "" && len(r.Commanders) > 0 {
		name = strings.Join(r.Commanders, " / ")
	}
	if name == "" && deckcoverage.IsListKey(r.DeckKey) {
		name = "Pasted list (" + r.DeckKey + ")"
	}
	if name == "" {
		name = r.DeckKey
	}
	return deckRequestTitlePrefix + plainUserText(name, deckRequestTitleMax)
}

// plainUserText is mdUserText for an issue TITLE, which GitHub does not
// render as markdown: escaping there would show the backslashes.
func plainUserText(s string, n int) string {
	s = redact.Secrets(s)
	s = strings.NewReplacer("\n", " ", "\r", " ").Replace(s)
	s = clip(s, n)
	s = strings.ReplaceAll(s, "@", "@​")
	return strings.TrimSpace(s)
}

// deckBucketLabels are the words a published report uses per bucket.
var deckBucketLabels = map[deckcoverage.Bucket]string{
	deckcoverage.Manual:     "Not in the engine (manual)",
	deckcoverage.Unreviewed: "In the engine, not yet reviewed",
	deckcoverage.Caveats:    "Automated with caveats",
	deckcoverage.Automated:  "Automated",
	deckcoverage.NoEffect:   "Nothing to automate",
}

// renderDeckRequestIssue is ADR 0095 §3 step 4's body: the deck link,
// who asked (display name only — never the snowflake), the counts, a
// checklist of the cards to add and to review, the caveated cards, and
// which surface filed it. Checklist entries are card names off the
// Scryfall index and caveats out of this repository; the deck name,
// the requester's name and any unresolved name came from outside and
// go through mdUserText.
func renderDeckRequestIssue(who deckRequester, fromBot bool, r *deckcoverage.Report) string {
	var b strings.Builder
	pasted := r.SourceURL == ""
	if pasted {
		// No link to show: say so, and put the list itself below.
		b.WriteString("**Deck:** Pasted list\n")
	} else {
		deckName := strings.TrimSpace(r.DeckName)
		if deckName == "" {
			deckName = "(unnamed deck)"
		}
		fmt.Fprintf(&b, "**Deck:** [%s](%s)\n", mdUserText(deckName, deckRequestTitleMax), redact.Secrets(r.SourceURL))
	}
	if len(r.Commanders) > 0 {
		fmt.Fprintf(&b, "**Commander:** %s\n", mdUserText(strings.Join(r.Commanders, " / "), deckRequestTitleMax))
	}
	fmt.Fprintf(&b, "**Requested by:** %s\n\n", mdUserText(who.DisplayName, deckRequesterNameMax))

	b.WriteString("| Cards | Distinct |\n|---|---:|\n")
	for _, bucket := range deckcoverage.Buckets {
		fmt.Fprintf(&b, "| %s | %d |\n", deckBucketLabels[bucket], r.Counts[bucket])
	}

	writeChecklist := func(heading string, cards []deckcoverage.Card) {
		if len(cards) == 0 {
			return
		}
		fmt.Fprintf(&b, "\n## %s\n\n", heading)
		for _, c := range cards {
			fmt.Fprintf(&b, "- [ ] %s (%s)\n", c.Name, c.OracleID)
		}
	}
	writeChecklist("Cards to add", r.CardsIn(deckcoverage.Manual))
	writeChecklist("Cards to review", r.CardsIn(deckcoverage.Unreviewed))

	if caveated := r.CardsIn(deckcoverage.Caveats); len(caveated) > 0 {
		b.WriteString("\n## Automated with caveats\n\n")
		for _, c := range caveated {
			fmt.Fprintf(&b, "- %s — %s\n", c.Name, strings.Join(c.Caveats, " "))
		}
	}
	if len(r.Unknown) > 0 {
		b.WriteString("\n## Names the card index could not resolve\n\n")
		for _, n := range r.Unknown {
			fmt.Fprintf(&b, "- %s\n", mdUserText(n, 120))
		}
	}
	if pasted {
		writePastedList(&b, r)
	}

	b.WriteString("\n---\n")
	if fromBot {
		b.WriteString("_Filed by `/c2-deck-req` in Discord (ADR 0095)._\n")
	} else {
		b.WriteString("_Filed from the deck checker on the site (ADR 0095)._\n")
	}
	return b.String()
}

// writePastedList appends a pasted list's cards as a collapsed block,
// so the issue carries the deck a link would have pointed at. It is
// rendered from the REPORT — card names off the Scryfall index, with
// their counts — and never from the raw paste: nothing the requester
// typed reaches this block (their unresolved names are in their own,
// escaped section above). A fenced block of "N Name" lines is also a
// list the checker reads back, with the commanders first and marked
// *CMDR*.
func writePastedList(b *strings.Builder, r *deckcoverage.Report) {
	commander := map[string]bool{}
	for _, n := range r.Commanders {
		commander[n] = true
	}
	cards := append([]deckcoverage.Card(nil), r.Cards...)
	sort.SliceStable(cards, func(i, j int) bool {
		if ci, cj := commander[cards[i].Name], commander[cards[j].Name]; ci != cj {
			return ci
		}
		return strings.ToLower(cards[i].Name) < strings.ToLower(cards[j].Name)
	})
	total := 0
	for _, c := range cards {
		total += c.Count
	}
	fmt.Fprintf(b, "\n<details>\n<summary>The list (%d cards)</summary>\n\n```text\n", total)
	for _, c := range cards {
		fmt.Fprintf(b, "%d %s", c.Count, c.Name)
		if commander[c.Name] {
			b.WriteString(" *CMDR*")
		}
		b.WriteString("\n")
	}
	b.WriteString("```\n\n</details>\n")
}

// renderDeckRequestComment is the "Also requested by" comment, with
// the counts as they are now: the deck may have changed since the
// issue was filed.
func renderDeckRequestComment(who deckRequester, r *deckcoverage.Report) string {
	var parts []string
	for _, bucket := range deckcoverage.Buckets {
		parts = append(parts, fmt.Sprintf("%s: %d", strings.ToLower(deckBucketLabels[bucket]), r.Counts[bucket]))
	}
	return fmt.Sprintf("Also requested by %s.\n\nCurrent counts — %s.\n",
		mdUserText(who.DisplayName, deckRequesterNameMax), strings.Join(parts, "; "))
}

// --- the report cache ---

// deckReportCache holds reports by deck key for a fixed TTL, bounded
// in size. A cached *Report is shared between responses and never
// modified after Build returns it.
type deckReportCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	max     int
	now     func() time.Time
	entries map[string]deckReportEntry
}

type deckReportEntry struct {
	report  *deckcoverage.Report
	expires time.Time
}

func newDeckReportCache(ttl time.Duration, max int, now func() time.Time) *deckReportCache {
	return &deckReportCache{ttl: ttl, max: max, now: now, entries: map[string]deckReportEntry{}}
}

func (c *deckReportCache) get(key string) (*deckcoverage.Report, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if !c.now().Before(e.expires) {
		delete(c.entries, key)
		return nil, false
	}
	return e.report, true
}

func (c *deckReportCache) put(key string, r *deckcoverage.Report) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if _, ok := c.entries[key]; !ok && len(c.entries) >= c.max {
		// Full: drop what has expired, then — if that freed nothing —
		// the entry closest to expiring.
		var oldestKey string
		var oldest time.Time
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
				continue
			}
			if oldestKey == "" || e.expires.Before(oldest) {
				oldestKey, oldest = k, e.expires
			}
		}
		if len(c.entries) >= c.max && oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}
	c.entries[key] = deckReportEntry{report: r, expires: now.Add(c.ttl)}
}
