package bot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// deckRequestCustomIDPrefix namespaces the /c2-deck-check "Request
// these cards" button's CustomIDs, the way c2EndCustomIDPrefix does
// for /c2-end's Confirm/Cancel buttons.
const deckRequestCustomIDPrefix = "c2-deck-req:"

// discordCustomIDMax is Discord's hard cap on a component CustomID.
const discordCustomIDMax = 100

// deckLinkTTL is how long a deck link stashed behind a short token
// (see buildDeckRequestCustomID) stays resolvable. In-memory only,
// same "no SIGHUP reload" trade-off end.go's confirmations accept: a
// bot restart drops it and the operator re-runs /c2-deck-check. An
// hour is generous headroom for someone to notice the button and
// click it without holding the token forever.
const deckLinkTTL = time.Hour

// errDeckLinkExpired is returned by resolveDeckRequestLink when a
// token custom ID names an entry that has aged out of the store.
var errDeckLinkExpired = errors.New("deck link no longer available")

// deckLinkEntry is one stashed deck link.
type deckLinkEntry struct {
	Link      string
	ExpiresAt time.Time
}

// deckLinkStore holds deck links behind a /c2-deck-check "Request
// these cards" button whose URL-encoded form doesn't fit Discord's
// 100-character CustomID cap. Keyed by a random token, the same
// pattern endConfirmations uses for /c2-end.
type deckLinkStore struct {
	mu      sync.Mutex
	entries map[string]deckLinkEntry
}

func newDeckLinkStore() *deckLinkStore {
	return &deckLinkStore{entries: make(map[string]deckLinkEntry)}
}

// put stores link under token, sweeping every already-expired entry
// along the way — same cheap, opportunistic cleanup endConfirmations.put
// does.
func (s *deckLinkStore) put(token, link string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.entries {
		if now.After(v.ExpiresAt) {
			delete(s.entries, k)
		}
	}
	s.entries[token] = deckLinkEntry{Link: link, ExpiresAt: now.Add(deckLinkTTL)}
}

func (s *deckLinkStore) get(token string, now time.Time) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[token]
	if !ok || now.After(e.ExpiresAt) {
		return "", false
	}
	return e.Link, true
}

// buildDeckRequestCustomID builds the "Request these cards" button's
// CustomID for link. It embeds the URL-encoded link directly when
// that fits Discord's 100-character cap (true for every Moxfield and
// Archidekt link seen so far); otherwise it stashes the link in store
// under a random token and embeds the token instead (#1631, ADR 0095 §4).
func buildDeckRequestCustomID(link string, store *deckLinkStore, now time.Time) string {
	direct := deckRequestCustomIDPrefix + "link:" + url.QueryEscape(link)
	if len(direct) <= discordCustomIDMax {
		return direct
	}
	token := randomToken()
	store.put(token, link, now)
	return deckRequestCustomIDPrefix + "tok:" + token
}

// resolveDeckRequestLink is buildDeckRequestCustomID's inverse: given
// a button's CustomID, recover the deck link it names.
func resolveDeckRequestLink(customID string, store *deckLinkStore, now time.Time) (string, error) {
	rest := strings.TrimPrefix(customID, deckRequestCustomIDPrefix)
	kind, val, ok := strings.Cut(rest, ":")
	if !ok || val == "" {
		return "", errors.New("malformed deck-request custom id")
	}
	switch kind {
	case "link":
		decoded, err := url.QueryUnescape(val)
		if err != nil {
			return "", fmt.Errorf("decode deck link: %w", err)
		}
		return decoded, nil
	case "tok":
		link, found := store.get(val, now)
		if !found {
			return "", errDeckLinkExpired
		}
		return link, nil
	default:
		return "", errors.New("malformed deck-request custom id")
	}
}

// --- /c2-deck-check ---

// handleDeckCheck runs /c2-deck-check: defer ephemerally (a deck
// fetch can run long), check coverage, and edit the deferred reply
// with the report — always ephemeral, per ADR 0095 §4.
func (h *Handler) handleDeckCheck(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	_ = s.InteractionRespond(i.Interaction, deferredEphemeralResponse())

	link := stringOption(data.Options, "link")
	report, err := h.checkDeck(ctx, link)
	if err != nil {
		h.log.Warn("deck-coverage failed", "error", err.Error(), "link", link)
		h.editEphemeral(s, i, deckCoverageErrorMessage(err), nil)
		return
	}

	content, components := deckCheckReply(report, h.cfg.ClientBaseURL, link, h.deckLinks, h.now())
	h.editEphemeral(s, i, content, components)
}

// checkDeck is /c2-deck-check's server call, split from the Discord
// response so it can be tested without a live session (the
// createInviteGame pattern).
func (h *Handler) checkDeck(ctx context.Context, link string) (DeckCoverageReport, error) {
	return h.client.DeckCoverage(ctx, link)
}

// deckCheckReply builds the full /c2-deck-check reply — content and,
// when the report has anything worth requesting, the "Request these
// cards" button — split out from handleDeckCheck so the assembly is
// testable without a live Discord session.
func deckCheckReply(report DeckCoverageReport, clientBaseURL, link string, store *deckLinkStore, now time.Time) (string, []discordgo.MessageComponent) {
	content := deckCheckContent(report, clientBaseURL, link)
	if !deckCheckWantsButton(report) {
		return content, nil
	}
	customID := buildDeckRequestCustomID(link, store, now)
	return content, []discordgo.MessageComponent{deckRequestButtonRow(customID)}
}

// deckCheckContent renders the ephemeral /c2-deck-check reply: the
// bucket counts in player words, up to ~15 manual card names, and a
// link to the site's full report.
func deckCheckContent(report DeckCoverageReport, clientBaseURL, link string) string {
	var b strings.Builder
	name := report.DeckName
	if name == "" {
		name = "This deck"
	}
	fmt.Fprintf(&b, "**%s**\n", name)
	fmt.Fprintf(&b, "%d need manual play, %d are unreviewed, %d work with caveats, %d are fully automated, %d need no automation.\n",
		report.Counts[BucketManual], report.Counts[BucketUnreviewed], report.Counts[BucketCaveats],
		report.Counts[BucketAutomated], report.Counts[BucketNoEffect])
	if names := manualCardNamesLine(report, 15); names != "" {
		b.WriteString(names)
		b.WriteString("\n")
	}
	b.WriteString(fullReportURL(clientBaseURL, link))
	return b.String()
}

// manualCardNamesLine lists up to max manual-bucket card names,
// "…and N more" past that. Empty when there are none.
func manualCardNamesLine(report DeckCoverageReport, max int) string {
	var names []string
	for _, c := range report.Cards {
		if c.Bucket == BucketManual {
			names = append(names, c.Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	shown := names
	suffix := ""
	if len(names) > max {
		shown = names[:max]
		suffix = fmt.Sprintf(", …and %d more", len(names)-max)
	}
	return "Manual: " + strings.Join(shown, ", ") + suffix
}

// fullReportURL builds the link to the site's full deck-check report.
func fullReportURL(clientBaseURL, link string) string {
	return fmt.Sprintf("Full report: %s/#/deck-check?url=%s", strings.TrimRight(clientBaseURL, "/"), url.QueryEscape(link))
}

// deckCheckWantsButton reports whether the report has anything worth
// a "Request these cards" button — manual or unreviewed cards.
func deckCheckWantsButton(report DeckCoverageReport) bool {
	return report.Counts[BucketManual] > 0 || report.Counts[BucketUnreviewed] > 0
}

// deckRequestButtonRow is the "Request these cards" button, shared by
// /c2-deck-check's reply.
func deckRequestButtonRow(customID string) discordgo.ActionsRow {
	return discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Request these cards",
				Style:    discordgo.PrimaryButton,
				CustomID: customID,
			},
		},
	}
}

// deckCoverageErrorMessage maps a checkDeck error to a user-visible
// string. POST /deck-coverage's own error responses already carry a
// player-readable sentence (docs/lobby.md); everything else falls
// back to the same wording the other commands use.
func deckCoverageErrorMessage(err error) string {
	var apiErr *DeckAPIError
	if errors.As(err, &apiErr) && apiErr.Message != "" {
		return apiErr.Message
	}
	switch {
	case errors.Is(err, ErrServerUnreachable):
		return "Game server is not reachable right now."
	case errors.Is(err, ErrUnauthorized):
		return "Bot is not authorized against the game server — check CMDCTRL_ADMIN_TOKEN."
	default:
		return "Something went wrong checking that deck."
	}
}

// --- /c2-deck-req and the "Request these cards" button ---

// handleDeckReq runs /c2-deck-req.
func (h *Handler) handleDeckReq(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	link := stringOption(data.Options, "link")
	h.runDeckRequest(ctx, s, i, link)
}

// dispatchDeckRequestComponent handles a "Request these cards" button
// click, running the same request flow as /c2-deck-req for whoever
// pressed it (ADR 0095 §4).
func (h *Handler) dispatchDeckRequestComponent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	link, err := resolveDeckRequestLink(customID, h.deckLinks, h.now())
	if err != nil {
		h.log.Warn("deck-request button resolve failed", "error", err.Error(), "custom_id", customID)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(deckLinkResolveErrorMessage(err)))
		return
	}
	h.runDeckRequest(ctx, s, i, link)
}

// runDeckRequest is the shared body of /c2-deck-req and the deck-
// request button: defer ephemerally (a deck fetch plus a GitHub call
// can run long), file or join the request, and reply according to
// its status.
//
// A deferred reply's visibility is fixed the moment it is sent —
// discordgo.WebhookEdit carries no Flags field, so InteractionResponseEdit
// cannot turn an ephemeral deferred reply channel-visible. So every
// deck-request reply defers ephemerally, and a channel-visible result
// (filed, or joined without already_requested) deletes that ephemeral
// placeholder and posts the visible half as a follow-up message
// instead of an edit; an ephemeral result just edits the placeholder.
func (h *Handler) runDeckRequest(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, link string) {
	_ = s.InteractionRespond(i.Interaction, deferredEphemeralResponse())

	requester := deckRequester(i.Interaction)
	result, err := h.requestDeck(ctx, link, requester)
	if err != nil {
		h.log.Warn("deck-request failed", "error", err.Error(), "link", link)
		h.editEphemeral(s, i, deckRequestErrorMessage(err), nil)
		return
	}

	content, channelVisible := deckRequestOutcomeMessage(result, link)
	h.finishDeckRequest(s, i, content, channelVisible)
}

// requestDeck is /c2-deck-req's (and the button's) server call, split
// from the Discord response so it can be tested without a live
// session (the createInviteGame pattern).
func (h *Handler) requestDeck(ctx context.Context, link string, requester DeckRequester) (DeckRequestResult, error) {
	return h.client.DeckRequest(ctx, link, requester)
}

// deckRequester builds the {discord_id, display_name} POST
// /deck-requests wants for the bot's admin session: the display name
// prefers a guild nickname, then the Discord global (display) name,
// then the bare username (ADR 0095 §4).
func deckRequester(i *discordgo.Interaction) DeckRequester {
	id, _ := interactionInvoker(i)
	return DeckRequester{DiscordID: id, DisplayName: invokerDisplayName(i)}
}

// invokerDisplayName reads the invoker's Member.Nick, then
// User.GlobalName, then User.Username — in that order, the first
// non-empty one wins.
func invokerDisplayName(i *discordgo.Interaction) string {
	if i == nil {
		return ""
	}
	if i.Member != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		if i.Member.User != nil {
			return i.Member.User.DisplayName()
		}
	}
	if i.User != nil {
		return i.User.DisplayName()
	}
	return ""
}

// deckRequestOutcomeMessage maps a successful DeckRequest result to
// its reply text and whether that reply belongs in the channel
// (ADR 0095 §4):
//
//   - filed: channel-visible, with the issue link and how many cards.
//   - joined: channel-visible, unless already_requested (ephemeral).
//   - nothing_to_add: ephemeral, with the counts.
//   - rate_limited: ephemeral, with the wait in hours/minutes.
func deckRequestOutcomeMessage(result DeckRequestResult, link string) (string, bool) {
	switch result.Status {
	case DeckRequestFiled:
		return fmt.Sprintf("Requested! %s — %d cards to add", result.IssueURL, manualPlusUnreviewed(result.Report)), true
	case DeckRequestJoined:
		if result.AlreadyRequested {
			return fmt.Sprintf("You already asked for this deck: %s", link), false
		}
		return fmt.Sprintf("Added your request to %s", result.IssueURL), true
	case DeckRequestNothingToAdd:
		return fmt.Sprintf("Everything in this deck is already in the engine. %s", countsSentence(result.Report)), false
	case DeckRequestRateLimited:
		return fmt.Sprintf("You've already asked for a few decks today — try again in %s.", retryAfterWords(result.RetryAfter)), false
	default:
		return fmt.Sprintf("Something unexpected happened requesting %s.", link), false
	}
}

// manualPlusUnreviewed is "N cards to add" — the manual and
// unreviewed cards a filed issue's checklist covers.
func manualPlusUnreviewed(report *DeckCoverageReport) int {
	if report == nil {
		return 0
	}
	return report.Counts[BucketManual] + report.Counts[BucketUnreviewed]
}

// countsSentence renders the automated/caveats/no-effect counts for
// the nothing_to_add reply.
func countsSentence(report *DeckCoverageReport) string {
	if report == nil {
		return ""
	}
	return fmt.Sprintf("(%d automated, %d with caveats, %d need no automation)",
		report.Counts[BucketAutomated], report.Counts[BucketCaveats], report.Counts[BucketNoEffect])
}

// retryAfterWords renders a retry_after second count as hours and
// minutes, per ADR 0095 §4.
func retryAfterWords(seconds int) string {
	if seconds <= 0 {
		return "less than a minute"
	}
	d := time.Duration(seconds) * time.Second
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return "less than a minute"
	}
}

// deckRequestErrorMessage maps a requestDeck error to a user-visible
// string, using the server's own sentence where there is one. 503 —
// CMDCTRL_GITHUB_TOKEN or the database missing — gets a fixed,
// player-facing message instead of the server's operator-facing one
// (ADR 0095 §4).
func deckRequestErrorMessage(err error) string {
	var apiErr *DeckAPIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == http.StatusServiceUnavailable {
			return "Deck requests aren't set up on this server."
		}
		if apiErr.Message != "" {
			return apiErr.Message
		}
	}
	switch {
	case errors.Is(err, ErrServerUnreachable):
		return "Game server is not reachable right now."
	case errors.Is(err, ErrUnauthorized):
		return "Bot is not authorized against the game server — check CMDCTRL_ADMIN_TOKEN."
	default:
		return "Something went wrong talking to the game server."
	}
}

// deckLinkResolveErrorMessage maps a resolveDeckRequestLink error to
// a user-visible string for a stale or malformed button.
func deckLinkResolveErrorMessage(err error) string {
	if errors.Is(err, errDeckLinkExpired) {
		return "This button has expired. Run `/c2-deck-check` again."
	}
	return "Something went wrong with that button. Run `/c2-deck-check` again."
}

// --- deferred-reply plumbing shared by both deck commands ---

// deferredEphemeralResponse is the immediate ack both deck commands
// (and the deck-request button) send: acknowledge now, ephemerally,
// and answer for real once the server call finishes. This is the
// bot's first deferred reply (ADR 0095 §4) — every other command
// still answers within Discord's 3s deadline directly.
func deferredEphemeralResponse() *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}
}

// editEphemeral finishes a deferred reply in place — the deferred ack
// was already ephemeral, and a deferred reply's visibility can't
// change, so this is only ever used for an ephemeral result.
func (h *Handler) editEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string, components []discordgo.MessageComponent) {
	comps := components
	if comps == nil {
		comps = []discordgo.MessageComponent{}
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &comps,
	}); err != nil {
		h.log.Error("interaction response edit failed", "error", err.Error())
	}
}

// finishDeckRequest delivers a deck-request result. A channel-visible
// result deletes the ephemeral "thinking" placeholder and posts the
// visible content as a follow-up message (the only way to get a
// non-ephemeral reply out of an interaction that deferred
// ephemerally); an ephemeral result just edits the placeholder.
func (h *Handler) finishDeckRequest(s *discordgo.Session, i *discordgo.InteractionCreate, content string, channelVisible bool) {
	if !channelVisible {
		h.editEphemeral(s, i, content, nil)
		return
	}
	if err := s.InteractionResponseDelete(i.Interaction); err != nil {
		h.log.Warn("interaction response delete failed", "error", err.Error())
	}
	if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{Content: content}); err != nil {
		h.log.Error("deck-request followup failed", "error", err.Error())
	}
}
