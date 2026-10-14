package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// Card lookup (#3324): /c2-card and the [[Card Name]] chat watcher.
//
// Both read the PUBLIC catalog (GET /catalog, GET /catalog/image/{id}; ADR
// 0042, S70) over the bot's server connection and need no session. The
// catalog holds exactly the cards the engine registers, so it carries the
// printed text and art the site already serves publicly, and a lookup that
// finds nothing there answers "not in the catalog" rather than guessing from
// Scryfall: that keeps this feature inside what ADR 0042 and ADR 0092 already
// expose, and adds no new card-data source.

// CatalogCard is the subset of catalog.Entry the bot renders. Re-declared
// (as ListGames does for lobby.listResponse) so the bot does not import the
// catalog package, which would pull in the effects registry.
type CatalogCard struct {
	OracleID     string        `json:"oracle_id"`
	ScryfallID   string        `json:"scryfall_id"`
	Name         string        `json:"name"`
	ManaCost     string        `json:"mana_cost"`
	TypeLine     string        `json:"type_line"`
	OracleText   string        `json:"oracle_text"`
	Completeness string        `json:"completeness"`
	Caveats      []string      `json:"caveats"`
	Faces        []CatalogFace `json:"faces"`
	Missing      bool          `json:"missing"`
}

// CatalogFace is one printed half of a multi-faced catalog card.
type CatalogFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

// Catalog calls the public GET /catalog. No session is involved.
func (c *ServerClient) Catalog(ctx context.Context) ([]CatalogCard, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/catalog", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServerUnreachable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, statusErr(resp)
	}
	var body struct {
		Cards []CatalogCard `json:"cards"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}
	return body.Cards, nil
}

// catalogTTL matches the server's Cache-Control on GET /catalog: the body
// changes only on a deploy or a dump reload.
const catalogTTL = 5 * time.Minute

// cardLookup finds catalog cards by name. It holds the catalog in memory and
// refetches after catalogTTL; a failed refetch keeps serving the old copy.
type cardLookup struct {
	fetch func(context.Context) ([]CatalogCard, error)
	now   func() time.Time

	mu      sync.Mutex
	cards   []CatalogCard
	keys    []string // normalised names and face names, parallel to idx
	idx     []int    // index into cards for each key
	fetched time.Time
}

func newCardLookup(fetch func(context.Context) ([]CatalogCard, error), now func() time.Time) *cardLookup {
	return &cardLookup{fetch: fetch, now: now}
}

// normName lowercases and reduces a name to letters and digits separated by
// single spaces, so "Sol  Ring", "sol-ring" and "SOL RING" agree.
func normName(s string) string {
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && b.Len() > 0 {
				b.WriteByte(' ')
			}
			space = false
			b.WriteRune(r)
		} else {
			space = true
		}
	}
	return b.String()
}

// loadLocked refreshes the catalog when stale. Caller holds l.mu.
func (l *cardLookup) loadLocked(ctx context.Context) error {
	if l.cards != nil && l.now().Sub(l.fetched) < catalogTTL {
		return nil
	}
	cards, err := l.fetch(ctx)
	if err != nil {
		if l.cards != nil {
			return nil
		}
		return err
	}
	if cards == nil {
		cards = []CatalogCard{}
	}
	l.cards, l.keys, l.idx = cards, nil, nil
	for i, c := range cards {
		l.keys = append(l.keys, normName(c.Name))
		l.idx = append(l.idx, i)
		for _, f := range c.Faces {
			if f.Name != "" {
				l.keys = append(l.keys, normName(f.Name))
				l.idx = append(l.idx, i)
			}
		}
	}
	l.fetched = l.now()
	return nil
}

// maxSuggestions caps the "did you mean" names on a miss.
const maxSuggestions = 3

// Find resolves query to a catalog card. Order: exact name (or face name),
// then a unique name that starts with the query, then a unique name that
// contains it. Anything still ambiguous or unmatched returns no card and up
// to maxSuggestions candidate names.
func (l *cardLookup) Find(ctx context.Context, query string) (*CatalogCard, []string, error) {
	q := normName(query)
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.loadLocked(ctx); err != nil {
		return nil, nil, err
	}
	if q == "" {
		return nil, nil, nil
	}
	pick := func(match func(key string) bool) []int {
		seen := map[int]bool{}
		var out []int
		for k, key := range l.keys {
			if match(key) && !seen[l.idx[k]] {
				seen[l.idx[k]] = true
				out = append(out, l.idx[k])
			}
		}
		return out
	}
	for _, match := range []func(string) bool{
		func(k string) bool { return k == q },
		func(k string) bool { return strings.HasPrefix(k, q) },
		func(k string) bool { return strings.Contains(k, q) },
	} {
		hits := pick(match)
		if len(hits) == 1 {
			c := l.cards[hits[0]]
			return &c, nil, nil
		}
		if len(hits) > 1 {
			names := make([]string, 0, len(hits))
			for _, h := range hits {
				names = append(names, l.cards[h].Name)
			}
			sort.Strings(names)
			if len(names) > maxSuggestions {
				names = names[:maxSuggestions]
			}
			return nil, names, nil
		}
	}
	return nil, nil, nil
}

// supportLine says whether the engine supports the card (the issue's
// "Supported in CMD & CTRL" line).
func supportLine(c CatalogCard) string {
	switch c.Completeness {
	case "full":
		return "Supported in CMD & CTRL."
	case "caveats":
		s := "Supported in CMD & CTRL, with gaps"
		if len(c.Caveats) > 0 {
			s += ": " + strings.Join(c.Caveats, "; ")
		}
		return s + "."
	default:
		return "In the CMD & CTRL catalog; how faithfully it plays hasn't been reviewed yet."
	}
}

// notFoundLine is the miss reply. The catalog only lists cards the engine
// registers, so a miss means "not supported yet" for any real card.
func notFoundLine(name string, suggestions []string) string {
	s := fmt.Sprintf("Nothing called \"%s\" is in the CMD & CTRL catalog.", name)
	if len(suggestions) > 0 {
		s += " Did you mean: " + strings.Join(suggestions, ", ") + "?"
	}
	return s + " If it's a real card we don't support yet, request it with `/c2-deck-req`."
}

const maxCardTextLen = 1500

func cardText(c CatalogCard) string {
	text := c.OracleText
	if text == "" && len(c.Faces) > 0 {
		parts := make([]string, 0, len(c.Faces))
		for _, f := range c.Faces {
			parts = append(parts, fmt.Sprintf("**%s** (%s)\n%s", f.Name, f.TypeLine, f.OracleText))
		}
		text = strings.Join(parts, "\n\n")
	}
	if r := []rune(text); len(r) > maxCardTextLen {
		text = string(r[:maxCardTextLen]) + "..."
	}
	return text
}

// cardEmbed renders a catalog card: art, mana cost, type, printed text and
// the support line. clientBase is the public origin that serves
// /catalog/image/{id}, which Discord fetches itself.
func cardEmbed(c CatalogCard, clientBase string) *discordgo.MessageEmbed {
	title := c.Name
	if c.ManaCost != "" {
		title += "  " + c.ManaCost
	}
	var desc strings.Builder
	if c.TypeLine != "" {
		desc.WriteString("*" + c.TypeLine + "*\n\n")
	}
	desc.WriteString(cardText(c))
	e := &discordgo.MessageEmbed{
		Title:       title,
		Description: desc.String(),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{{Name: "In CMD & CTRL", Value: supportLine(c)}},
	}
	if c.ScryfallID != "" && !c.Missing {
		e.Image = &discordgo.MessageEmbedImage{
			URL: strings.TrimRight(clientBase, "/") + "/catalog/image/" + url.PathEscape(c.ScryfallID),
		}
	}
	return e
}

// handleCard runs /c2-card.
func (h *Handler) handleCard(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	_ = s.InteractionRespond(i.Interaction, h.cardResponse(ctx, stringOption(data.Options, "name")))
}

// cardResponse builds /c2-card's reply, split from the session call so it
// can be tested. A hit is channel-visible (the point is to show the card to
// the chat); a miss or an error is ephemeral.
func (h *Handler) cardResponse(ctx context.Context, name string) *discordgo.InteractionResponse {
	name = strings.TrimSpace(name)
	card, sugg, err := h.cards.Find(ctx, name)
	if err != nil {
		h.log.Error("card lookup failed", "error", err.Error())
		return ephemeralResponse(inviteErrorMessage(err))
	}
	if card == nil {
		return ephemeralResponse(notFoundLine(name, sugg))
	}
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{cardEmbed(*card, h.cfg.ClientBaseURL)}},
	}
}

// Chat watcher limits.
const (
	maxCardsPerMessage = 3
	maxQueryLen        = 80
	// replyWindow and maxRepliesPerWindow rate-limit replies per channel.
	replyWindow         = time.Minute
	maxRepliesPerWindow = 6
)

var cardRefRE = regexp.MustCompile(`\[\[([^\[\]\n]{1,80})\]\]`)

// extractCardRefs returns up to limit distinct [[...]] names in order.
func extractCardRefs(content string, limit int) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range cardRefRE.FindAllStringSubmatch(content, -1) {
		n := strings.TrimSpace(m[1])
		k := normName(n)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, n)
		if len(out) == limit {
			break
		}
	}
	return out
}

// channelLimiter allows at most maxRepliesPerWindow replies per channel in
// any replyWindow.
type channelLimiter struct {
	mu   sync.Mutex
	sent map[string][]time.Time
}

func (l *channelLimiter) allow(channel string, now time.Time, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.sent == nil {
		l.sent = map[string][]time.Time{}
	}
	var kept []time.Time
	for _, t := range l.sent[channel] {
		if now.Sub(t) < replyWindow {
			kept = append(kept, t)
		}
	}
	if len(kept)+n > maxRepliesPerWindow {
		l.sent[channel] = kept
		return false
	}
	for j := 0; j < n; j++ {
		kept = append(kept, now)
	}
	l.sent[channel] = kept
	return true
}

// messageSender is the slice of *discordgo.Session the watcher replies with.
type messageSender interface {
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

// HandleMessage is the MessageCreate callback. It does nothing unless the
// channel is on CMDCTRL_DISCORD_CARD_CHANNEL_IDS in an allowed guild.
func (h *Handler) HandleMessage(s messageSender, m *discordgo.MessageCreate) {
	if m == nil || m.Message == nil || m.Author == nil || m.Author.Bot {
		return
	}
	if !h.cfg.GuildAllowed(m.GuildID) || !h.cfg.CardChannelAllowed(m.ChannelID) {
		return
	}
	refs := extractCardRefs(m.Content, maxCardsPerMessage)
	if len(refs) == 0 {
		return
	}
	if !h.cardLimiter.allow(m.ChannelID, h.now(), len(refs)) {
		h.log.Info("card lookup rate-limited", "channel", m.ChannelID)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultInteractionTimeout)
	defer cancel()
	for _, name := range refs {
		card, sugg, err := h.cards.Find(ctx, name)
		if err != nil {
			h.log.Error("card lookup failed", "error", err.Error())
			return
		}
		msg := &discordgo.MessageSend{
			Reference:       m.Reference(),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		}
		if card == nil {
			msg.Content = notFoundLine(name, sugg)
		} else {
			msg.Embeds = []*discordgo.MessageEmbed{cardEmbed(*card, h.cfg.ClientBaseURL)}
		}
		if _, err := s.ChannelMessageSendComplex(m.ChannelID, msg); err != nil {
			h.log.Warn("card reply failed", "error", err.Error())
			return
		}
	}
}
