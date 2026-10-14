package bot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

var testCatalog = []CatalogCard{
	{OracleID: "o1", ScryfallID: "11111111-1111-1111-1111-111111111111", Name: "Sol Ring", ManaCost: "{1}",
		TypeLine: "Artifact", OracleText: "{T}: Add {C}{C}.", Completeness: "full"},
	{OracleID: "o2", ScryfallID: "22222222-2222-2222-2222-222222222222", Name: "Sol Talisman", TypeLine: "Artifact",
		OracleText: "x", Completeness: "caveats", Caveats: []string{"ignores the damage clause"}},
	{OracleID: "o3", ScryfallID: "33333333-3333-3333-3333-333333333333", Name: "Witch Enchanter // Witch-Blessed Meadow",
		Completeness: "unreviewed", Faces: []CatalogFace{
			{Name: "Witch Enchanter", TypeLine: "Creature", OracleText: "front"},
			{Name: "Witch-Blessed Meadow", TypeLine: "Land", OracleText: "back"},
		}},
}

func testHandler(cfg Config, cards []CatalogCard, fetchErr error) (*Handler, *int) {
	calls := 0
	h := NewHandler(cfg, NewServerClient("http://unused", "x"), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	h.cards = newCardLookup(func(context.Context) ([]CatalogCard, error) {
		calls++
		return cards, fetchErr
	}, func() time.Time { return h.now() })
	return h, &calls
}

func TestCardFind(t *testing.T) {
	h, _ := testHandler(Config{}, testCatalog, nil)
	ctx := context.Background()
	for _, tc := range []struct {
		q, want string
		sugg    int
	}{
		{"sol ring", "Sol Ring", 0},
		{"  SOL-RING ", "Sol Ring", 0},
		{"ring", "Sol Ring", 0},
		{"witch-blessed meadow", "Witch Enchanter // Witch-Blessed Meadow", 0},
		{"sol", "", 2},
		{"black lotus", "", 0},
		{"!!", "", 0},
	} {
		c, sugg, err := h.cards.Find(ctx, tc.q)
		if err != nil {
			t.Fatalf("%q: %v", tc.q, err)
		}
		got := ""
		if c != nil {
			got = c.Name
		}
		if got != tc.want || len(sugg) != tc.sugg {
			t.Errorf("%q: got %q %v, want %q (%d suggestions)", tc.q, got, sugg, tc.want, tc.sugg)
		}
	}
}

func TestCardLookupCachesAndKeepsStaleOnError(t *testing.T) {
	now := time.Unix(1000, 0)
	calls, fail := 0, false
	l := newCardLookup(func(context.Context) ([]CatalogCard, error) {
		calls++
		if fail {
			return nil, ErrServerUnreachable
		}
		return testCatalog, nil
	}, func() time.Time { return now })
	ctx := context.Background()
	_, _, _ = l.Find(ctx, "sol ring")
	_, _, _ = l.Find(ctx, "sol ring")
	if calls != 1 {
		t.Fatalf("fetched %d times within the TTL, want 1", calls)
	}
	now = now.Add(catalogTTL + time.Second)
	fail = true
	c, _, err := l.Find(ctx, "sol ring")
	if err != nil || c == nil {
		t.Fatalf("stale copy should keep serving: %v %v", c, err)
	}

	cold := newCardLookup(func(context.Context) ([]CatalogCard, error) { return nil, ErrServerUnreachable }, time.Now)
	if _, _, err := cold.Find(ctx, "x"); !errors.Is(err, ErrServerUnreachable) {
		t.Errorf("cold failure: %v", err)
	}
}

func TestCatalogClientIsPublic(t *testing.T) {
	var auth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/catalog" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"counts": map[string]int{}, "cards": testCatalog})
	}))
	t.Cleanup(ts.Close)
	cards, err := NewServerClient(ts.URL, "secret").Catalog(context.Background())
	if err != nil || len(cards) != 3 || cards[0].Name != "Sol Ring" {
		t.Fatalf("Catalog: %v %v", cards, err)
	}
	if auth != "" {
		t.Errorf("catalog is public; sent Authorization %q", auth)
	}
}

func TestCardEmbed(t *testing.T) {
	e := cardEmbed(testCatalog[0], "https://cmd.labxp.io/")
	if e.Title != "Sol Ring  {1}" || !strings.Contains(e.Description, "Add {C}{C}") || !strings.Contains(e.Description, "Artifact") {
		t.Errorf("embed: %+v", e)
	}
	if e.Image == nil || e.Image.URL != "https://cmd.labxp.io/catalog/image/11111111-1111-1111-1111-111111111111" {
		t.Errorf("image: %+v", e.Image)
	}
	if got := e.Fields[0].Value; got != "Supported in CMD & CTRL." {
		t.Errorf("support: %q", got)
	}
	if got := supportLine(testCatalog[1]); !strings.Contains(got, "ignores the damage clause") {
		t.Errorf("caveats: %q", got)
	}
	faces := cardEmbed(testCatalog[2], "https://x")
	if !strings.Contains(faces.Description, "front") || !strings.Contains(faces.Description, "back") {
		t.Errorf("faces: %q", faces.Description)
	}
	missing := testCatalog[0]
	missing.Missing = true
	if cardEmbed(missing, "https://x").Image != nil {
		t.Error("a card with no printing must not link an image")
	}
}

func TestCardResponse(t *testing.T) {
	h, _ := testHandler(Config{ClientBaseURL: "https://x"}, testCatalog, nil)
	hit := h.cardResponse(context.Background(), " Sol Ring ")
	if hit.Data.Flags != 0 || len(hit.Data.Embeds) != 1 {
		t.Errorf("hit should be a visible embed: %+v", hit.Data)
	}
	miss := h.cardResponse(context.Background(), "Black Lotus")
	if miss.Data.Flags != discordgo.MessageFlagsEphemeral || !strings.Contains(miss.Data.Content, "/c2-deck-req") {
		t.Errorf("miss: %+v", miss.Data)
	}
	hErr, _ := testHandler(Config{}, nil, ErrServerUnreachable)
	if r := hErr.cardResponse(context.Background(), "x"); r.Data.Flags != discordgo.MessageFlagsEphemeral {
		t.Errorf("error should be ephemeral: %+v", r.Data)
	}
}

func TestExtractCardRefs(t *testing.T) {
	got := extractCardRefs("play [[Sol Ring]] and [[sol ring]] then [[A]] [[B]] [[C]] [[ ]] [[x\ny]]", 3)
	if strings.Join(got, "|") != "Sol Ring|A|B" {
		t.Errorf("got %v", got)
	}
	if got := extractCardRefs("no refs [here]", 3); got != nil {
		t.Errorf("got %v", got)
	}
}

type sentMsg struct {
	channel string
	msg     *discordgo.MessageSend
}

type fakeSender struct{ sent []sentMsg }

func (f *fakeSender) ChannelMessageSendComplex(ch string, d *discordgo.MessageSend, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	f.sent = append(f.sent, sentMsg{ch, d})
	return &discordgo.Message{}, nil
}

func chatMsg(guild, channel, content string, bot bool) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{Message: &discordgo.Message{
		ID: "m1", GuildID: guild, ChannelID: channel, Content: content,
		Author: &discordgo.User{ID: "u1", Bot: bot},
	}}
}

func TestHandleMessage(t *testing.T) {
	cfg := Config{GuildIDs: []string{"g1"}, CardChannelIDs: []string{"c1"}, ClientBaseURL: "https://x"}
	h, calls := testHandler(cfg, testCatalog, nil)
	s := &fakeSender{}

	h.HandleMessage(s, chatMsg("g1", "c1", "[[Sol Ring]] and [[Black Lotus]]", false))
	if len(s.sent) != 2 {
		t.Fatalf("want 2 replies, got %d", len(s.sent))
	}
	if len(s.sent[0].msg.Embeds) != 1 || s.sent[0].msg.Reference == nil || s.sent[0].msg.AllowedMentions == nil {
		t.Errorf("hit reply: %+v", s.sent[0].msg)
	}
	if !strings.Contains(s.sent[1].msg.Content, "Black Lotus") {
		t.Errorf("miss reply: %+v", s.sent[1].msg)
	}

	before := len(s.sent)
	h.HandleMessage(s, chatMsg("g1", "other", "[[Sol Ring]]", false))
	h.HandleMessage(s, chatMsg("g2", "c1", "[[Sol Ring]]", false))
	h.HandleMessage(s, chatMsg("g1", "c1", "[[Sol Ring]]", true))
	h.HandleMessage(s, chatMsg("g1", "c1", "no card reference", false))
	if len(s.sent) != before {
		t.Errorf("ignored messages got %d replies", len(s.sent)-before)
	}
	if *calls != 1 {
		t.Errorf("catalog fetched %d times", *calls)
	}
}

func TestHandleMessageRateLimitsPerChannel(t *testing.T) {
	cfg := Config{GuildIDs: []string{"g1"}, CardChannelIDs: []string{"c1", "c2"}}
	h, _ := testHandler(cfg, testCatalog, nil)
	now := time.Unix(5000, 0)
	h.now = func() time.Time { return now }
	s := &fakeSender{}
	for i := 0; i < 10; i++ {
		h.HandleMessage(s, chatMsg("g1", "c1", "[[Sol Ring]]", false))
	}
	if len(s.sent) != maxRepliesPerWindow {
		t.Fatalf("c1 got %d replies, want %d", len(s.sent), maxRepliesPerWindow)
	}
	h.HandleMessage(s, chatMsg("g1", "c2", "[[Sol Ring]]", false))
	if len(s.sent) != maxRepliesPerWindow+1 {
		t.Error("another channel has its own budget")
	}
	now = now.Add(replyWindow + time.Second)
	h.HandleMessage(s, chatMsg("g1", "c1", "[[Sol Ring]]", false))
	if len(s.sent) != maxRepliesPerWindow+2 {
		t.Error("the budget should refill after the window")
	}
}

func TestCardChannelConfig(t *testing.T) {
	t.Setenv("CMDCTRL_DISCORD_CARD_CHANNEL_IDS", " 1, 2 ,")
	cfg := ConfigFromEnv()
	if !cfg.CardChannelAllowed("2") || cfg.CardChannelAllowed("3") {
		t.Errorf("channels: %v", cfg.CardChannelIDs)
	}
	if (Config{}).CardChannelAllowed("1") {
		t.Error("empty list must watch nothing")
	}
}
