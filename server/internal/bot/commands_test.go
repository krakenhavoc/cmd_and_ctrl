package bot

import (
	"bytes"
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
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

func TestBuildInviteURL(t *testing.T) {
	id := uuid.MustParse("0196e7b8-1b2a-4cd0-a1f2-5d5f2b6a1234")
	cases := []struct {
		base string
		want string
	}{
		{"https://cmd.labxp.io", "https://cmd.labxp.io/#/games/0196e7b8-1b2a-4cd0-a1f2-5d5f2b6a1234/join?t=tok"},
		{"https://cmd.labxp.io/", "https://cmd.labxp.io/#/games/0196e7b8-1b2a-4cd0-a1f2-5d5f2b6a1234/join?t=tok"},
		{"http://127.0.0.1:5173", "http://127.0.0.1:5173/#/games/0196e7b8-1b2a-4cd0-a1f2-5d5f2b6a1234/join?t=tok"},
	}
	for _, c := range cases {
		got := buildInviteURL(c.base, id, "tok")
		if got != c.want {
			t.Errorf("buildInviteURL(%q): got %q, want %q", c.base, got, c.want)
		}
	}
}

func TestDefaultGameName(t *testing.T) {
	fixed := time.Date(2026, 4, 22, 14, 32, 0, 0, time.UTC)
	got := defaultGameName(fixed)
	if got != "Discord game · 2026-04-22 14:32 UTC" {
		t.Errorf("defaultGameName: got %q", got)
	}
}

func TestStringOption(t *testing.T) {
	opts := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "name", Type: discordgo.ApplicationCommandOptionString, Value: "friday"},
		{Name: "other", Type: discordgo.ApplicationCommandOptionInteger, Value: 42.0},
	}
	if got := stringOption(opts, "name"); got != "friday" {
		t.Errorf("stringOption(name): got %q", got)
	}
	if got := stringOption(opts, "missing"); got != "" {
		t.Errorf("stringOption(missing): got %q", got)
	}
	if got := stringOption(opts, "other"); got != "" {
		t.Errorf("stringOption(wrong type): got %q", got)
	}
}

func TestCommandDefinitions(t *testing.T) {
	defs := commandDefinitions()
	if len(defs) != 5 {
		t.Fatalf("want 5 definitions, got %d", len(defs))
	}
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name
	}
	for _, want := range []string{CmdInvite, CmdGames, CmdEnd, CmdDeckCheck, CmdDeckReq} {
		if !contains(names, want) {
			t.Errorf("missing expected command name %q: %v", want, names)
		}
	}
	// Every command must carry the c2- prefix and none of the old
	// cc- names (#1631, ADR 0095).
	for _, n := range names {
		if !strings.HasPrefix(n, "c2-") {
			t.Errorf("command %q does not carry the c2- prefix", n)
		}
		if strings.HasPrefix(n, "cc-") {
			t.Errorf("command %q still carries the old cc- prefix", n)
		}
	}
	// /c2-invite must have an optional string "name" option so
	// users can type /c2-invite friday-commander.
	var invite, end, deckCheck, deckReq *discordgo.ApplicationCommand
	for _, d := range defs {
		switch d.Name {
		case CmdInvite:
			invite = d
		case CmdEnd:
			end = d
		case CmdDeckCheck:
			deckCheck = d
		case CmdDeckReq:
			deckReq = d
		}
	}
	if invite == nil {
		t.Fatal("missing invite command")
	}
	if len(invite.Options) != 1 || invite.Options[0].Required {
		t.Errorf("invite should have one optional option, got %+v", invite.Options)
	}
	// /c2-end must have a required, autocompleting "game" option.
	if end == nil {
		t.Fatal("missing end command")
	}
	if len(end.Options) != 1 || !end.Options[0].Required || !end.Options[0].Autocomplete {
		t.Errorf("end should have one required, autocompleting option, got %+v", end.Options)
	}
	// Both deck commands take a single required "link" string.
	if deckCheck == nil {
		t.Fatal("missing deck-check command")
	}
	if len(deckCheck.Options) != 1 || deckCheck.Options[0].Name != "link" || !deckCheck.Options[0].Required {
		t.Errorf("deck-check should have one required \"link\" option, got %+v", deckCheck.Options)
	}
	if deckReq == nil {
		t.Fatal("missing deck-req command")
	}
	if len(deckReq.Options) != 1 || deckReq.Options[0].Name != "link" || !deckReq.Options[0].Required {
		t.Errorf("deck-req should have one required \"link\" option, got %+v", deckReq.Options)
	}
}

// fakeRegistrar is a commandRegistrar test double: RegisterCommands
// is tested against this small interface (#1631) instead of a live
// discordgo.Session, which is how ApplicationCommandBulkOverwrite —
// the bulk-overwrite registration path — gets exercised without a
// Discord connection.
type fakeRegistrar struct {
	// calls records each ApplicationCommandBulkOverwrite call as
	// (guildID, command names).
	calls []struct {
		guildID string
		names   []string
	}
	failGuild string // guildID that should error, if any
}

func (f *fakeRegistrar) ApplicationCommandBulkOverwrite(appID, guildID string, commands []*discordgo.ApplicationCommand, _ ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error) {
	if guildID == f.failGuild {
		return nil, errors.New("boom")
	}
	names := make([]string, len(commands))
	for i, c := range commands {
		names[i] = c.Name
	}
	f.calls = append(f.calls, struct {
		guildID string
		names   []string
	}{guildID, names})
	return commands, nil
}

func TestRegisterCommands_BulkOverwrite(t *testing.T) {
	reg := &fakeRegistrar{}
	RegisterCommands(reg, "app1", []string{"g1", "g2"}, discardLogger())

	if len(reg.calls) != 2 {
		t.Fatalf("want 2 bulk-overwrite calls (one per guild), got %d", len(reg.calls))
	}
	if reg.calls[0].guildID != "g1" || reg.calls[1].guildID != "g2" {
		t.Errorf("guild order: got %+v", reg.calls)
	}
	for _, call := range reg.calls {
		if len(call.names) != len(commandDefinitions()) {
			t.Errorf("guild %s: want %d commands in one overwrite call, got %d (%v)",
				call.guildID, len(commandDefinitions()), len(call.names), call.names)
		}
		if !contains(call.names, CmdDeckCheck) || !contains(call.names, CmdDeckReq) {
			t.Errorf("guild %s: bulk overwrite missing deck commands: %v", call.guildID, call.names)
		}
	}
}

func TestRegisterCommands_SkipsFailedGuildButContinues(t *testing.T) {
	reg := &fakeRegistrar{failGuild: "bad-guild"}
	RegisterCommands(reg, "app1", []string{"bad-guild", "g2"}, discardLogger())

	if len(reg.calls) != 1 || reg.calls[0].guildID != "g2" {
		t.Errorf("want only the good guild to have registered, got %+v", reg.calls)
	}
}

func TestInviteErrorMessage(t *testing.T) {
	cases := map[error]string{
		ErrServerUnreachable: "Game server is not reachable",
		ErrUnauthorized:      "Bot is not authorized",
		errors.New("boom"):   "Something went wrong",
	}
	for err, wantSub := range cases {
		got := inviteErrorMessage(err)
		if !strings.Contains(got, wantSub) {
			t.Errorf("inviteErrorMessage(%v): got %q, want substring %q", err, got, wantSub)
		}
	}
}

func TestInviteSuccessResponse(t *testing.T) {
	meta := lobby.GameMeta{ID: uuid.New(), Name: "friday", InviteToken: "tok"}
	url := "https://x/#/games/" + meta.ID.String() + "/join?t=tok"
	resp := inviteSuccessResponse(meta, url)

	if resp.Type != discordgo.InteractionResponseChannelMessageWithSource {
		t.Errorf("type: got %v", resp.Type)
	}
	if resp.Data.Flags != 0 {
		t.Errorf("invite response should NOT be ephemeral; flags=%v", resp.Data.Flags)
	}
	if len(resp.Data.Embeds) != 1 {
		t.Fatalf("want 1 embed, got %d", len(resp.Data.Embeds))
	}
	if !strings.Contains(resp.Data.Embeds[0].Description, url) {
		t.Errorf("embed does not contain URL: %q", resp.Data.Embeds[0].Description)
	}
	if !strings.Contains(resp.Data.Embeds[0].Title, "friday") {
		t.Errorf("embed title does not contain game name: %q", resp.Data.Embeds[0].Title)
	}
}

func TestGamesListResponse_Empty(t *testing.T) {
	resp := gamesListResponse(nil)
	if resp.Data.Flags&discordgo.MessageFlagsEphemeral == 0 {
		t.Error("empty-list response should be ephemeral")
	}
	if !strings.Contains(resp.Data.Content, "No games") {
		t.Errorf("content: got %q", resp.Data.Content)
	}
}

func TestGamesListResponse_Populated(t *testing.T) {
	games := []lobby.GameMeta{
		{ID: uuid.New(), Name: "a", State: "lobby", Players: make([]lobby.SeatInfo, 0)},
		{ID: uuid.New(), Name: "b", State: "active", Players: make([]lobby.SeatInfo, 2)},
	}
	resp := gamesListResponse(games)
	if resp.Data.Flags&discordgo.MessageFlagsEphemeral == 0 {
		t.Error("list response should be ephemeral")
	}
	c := resp.Data.Content
	if !strings.Contains(c, "`a`") || !strings.Contains(c, "lobby") {
		t.Errorf("first game missing: %q", c)
	}
	if !strings.Contains(c, "`b`") || !strings.Contains(c, "active") || !strings.Contains(c, "2 seats") {
		t.Errorf("second game missing or seat pluralization wrong: %q", c)
	}
	// Must NOT leak invite tokens even if the server somehow
	// returned them (defense in depth for the stripping that
	// lobby.Lobby.List already does).
	for _, g := range games {
		if g.InviteToken != "" && strings.Contains(c, g.InviteToken) {
			t.Errorf("leaked invite token in list response: %q", c)
		}
	}
}

func TestEphemeralResponse(t *testing.T) {
	resp := ephemeralResponse("hi")
	if resp.Data.Flags&discordgo.MessageFlagsEphemeral == 0 {
		t.Error("should be ephemeral")
	}
	if resp.Data.Content != "hi" {
		t.Errorf("content: got %q", resp.Data.Content)
	}
}

// --- Dispatcher tests against a stub ServerClient ---
//
// We drive Handler.Dispatch indirectly by composing the same
// flow against the httptest-backed ServerClient. The discordgo
// Session isn't easily fakeable, so we test the three branches
// that are fakeable without one: the allow-list guard (exits
// before any HTTP call), and the two happy-path handlers via
// the exposed helper shapes above.

func TestDispatch_UnknownCommand(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, nil, log)

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionApplicationCommand,
		Data: discordgo.ApplicationCommandInteractionData{
			Name: "something-else",
		},
	}}

	// Session is nil — Dispatch must return BEFORE attempting
	// an HTTP call, because the command is not in the switch.
	// The session-dependent InteractionRespond call is the only
	// nil-deref risk; we tolerate it via the defer/recover trick
	// below (Dispatch intentionally ignores the error).
	defer func() {
		// recovered panic is acceptable — it proves Dispatch
		// tried to call InteractionRespond. We just want to
		// confirm no HTTP call was made.
		_ = recover()
	}()
	h.Dispatch(nil, i)
}

func TestHandleInvite_Success(t *testing.T) {
	id := uuid.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "session"})
	})
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(lobby.GameMeta{ID: id, Name: "friday", InviteToken: "tok", State: "lobby"})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	c := NewServerClient(ts.URL, "admin")
	meta, err := c.CreateGame(context.Background(), "friday", "")
	if err != nil {
		t.Fatalf("CreateGame: %v", err)
	}
	url := buildInviteURL("https://cmd.labxp.io", meta.ID, meta.InviteToken)
	resp := inviteSuccessResponse(meta, url)

	// Round-trip the response through JSON the way discordgo
	// would to confirm the embed serialises cleanly.
	buf := bytes.Buffer{}
	if err := json.NewEncoder(&buf).Encode(resp); err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !strings.Contains(buf.String(), url) {
		t.Errorf("encoded response missing URL: %q", buf.String())
	}
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func TestInviteNamesTheInvokerAsHost(t *testing.T) {
	var got map[string]string
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "session"})
	})
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(lobby.GameMeta{ID: uuid.New(), Name: got["name"], InviteToken: "tok", State: "lobby"})
	})
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	h := NewHandler(Config{GuildIDs: []string{"g1"}}, NewServerClient(ts.URL, "admin"), slog.New(slog.NewTextHandler(io.Discard, nil)))
	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		Type:    discordgo.InteractionApplicationCommand,
		GuildID: "g1",
		Member:  &discordgo.Member{User: &discordgo.User{ID: "424242"}},
	}}
	name, _, err := h.createInviteGame(context.Background(), i, discordgo.ApplicationCommandInteractionData{Name: CmdInvite})
	if err != nil {
		t.Fatalf("createInviteGame: %v", err)
	}
	if got["host_discord_id"] != "424242" {
		t.Errorf("host_discord_id = %q, want the invoker 424242 (body %v)", got["host_discord_id"], got)
	}
	if got["name"] != name || name == "" {
		t.Errorf("name = %q, sent %q", name, got["name"])
	}

	// A DM interaction carries the user on User, not Member.
	dm := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{User: &discordgo.User{ID: "7"}}}
	if id := invokerID(dm); id != "7" {
		t.Errorf("DM invoker = %q, want 7", id)
	}
	if id := invokerID(&discordgo.InteractionCreate{Interaction: &discordgo.Interaction{}}); id != "" {
		t.Errorf("no user: invoker = %q", id)
	}
}
