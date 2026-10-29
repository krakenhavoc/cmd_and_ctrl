package bot

// Tests for bansync.go (ADR 0156's 2026-10-29 amendment, #3325): which
// gateway events ask the server to end a person's sessions, the intents
// ban sync asks for, the fallback when Discord refuses Server Members
// (4014), and the call the bot makes. Who has an account, who is exempt
// and whether a person is already out are the server's decisions,
// tested in lobby/ban_sync_test.go.

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/websocket"
)

const (
	syncGuild  = "100000000000000001"
	otherGuild = "100000000000000002"
	syncRole   = "200000000000000001"
	otherRole  = "200000000000000002"
	memberID   = "300000000000000001"
	adminID    = "300000000000000009"
)

func banSyncConfig() Config {
	return Config{
		GuildIDs:         []string{syncGuild, otherGuild},
		BanSync:          true,
		CommunityGuildID: syncGuild,
		PlayerRoleID:     syncRole,
		AdminUserIDs:     []string{adminID},
	}
}

func guildMember(guild, id string, roles ...string) *discordgo.Member {
	return &discordgo.Member{GuildID: guild, User: &discordgo.User{ID: id}, Roles: roles}
}

func TestBanSyncDecisions(t *testing.T) {
	on := banSyncConfig()
	off := banSyncConfig()
	off.BanSync = false

	ban := func(guild, id string) func(Config) (string, string) {
		return func(c Config) (string, string) {
			return c.banSyncReasonForBan(&discordgo.GuildBanAdd{GuildID: guild, User: &discordgo.User{ID: id}})
		}
	}
	update := func(m, before *discordgo.Member) func(Config) (string, string) {
		return func(c Config) (string, string) {
			return c.banSyncReasonForUpdate(&discordgo.GuildMemberUpdate{Member: m, BeforeUpdate: before})
		}
	}
	remove := func(guild, id string) func(Config) (string, string) {
		return func(c Config) (string, string) {
			return c.banSyncReasonForRemove(&discordgo.GuildMemberRemove{Member: guildMember(guild, id)})
		}
	}
	bot := &discordgo.GuildBanAdd{GuildID: syncGuild, User: &discordgo.User{ID: memberID, Bot: true}}

	for _, tc := range []struct {
		name string
		cfg  Config
		ask  func(Config) (string, string)
		want string
	}{
		{"a ban", on, ban(syncGuild, memberID), banSyncBanned},
		{"a ban in an unrelated guild", on, ban(otherGuild, memberID), ""},
		{"a ban with ban sync off", off, ban(syncGuild, memberID), ""},
		{"a ban of a bot", on, func(c Config) (string, string) { return c.banSyncReasonForBan(bot) }, ""},
		{"a ban of someone on the admin list", on, ban(syncGuild, adminID), ""},

		{"the role dropped", on, update(guildMember(syncGuild, memberID, otherRole), guildMember(syncGuild, memberID, syncRole, otherRole)), banSyncRoleRemoved},
		{"the role gone, nothing cached", on, update(guildMember(syncGuild, memberID, otherRole), nil), banSyncRoleRemoved},
		{"the role kept", on, update(guildMember(syncGuild, memberID, syncRole), guildMember(syncGuild, memberID, syncRole, otherRole)), ""},
		{"never had the role (a nickname change)", on, update(guildMember(syncGuild, memberID, otherRole), guildMember(syncGuild, memberID, otherRole)), ""},
		{"the role dropped in an unrelated guild", on, update(guildMember(otherGuild, memberID), guildMember(otherGuild, memberID, syncRole)), ""},
		{"the role dropped with ban sync off", off, update(guildMember(syncGuild, memberID), guildMember(syncGuild, memberID, syncRole)), ""},
		{"an admin's role dropped", on, update(guildMember(syncGuild, adminID), guildMember(syncGuild, adminID, syncRole)), ""},

		{"a leave or a kick", on, remove(syncGuild, memberID), banSyncLeft},
		{"a leave in an unrelated guild", on, remove(otherGuild, memberID), ""},
		{"a leave with ban sync off", off, remove(syncGuild, memberID), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, reason := tc.ask(tc.cfg)
			if reason != tc.want {
				t.Fatalf("reason = %q, want %q", reason, tc.want)
			}
			if tc.want != "" && id == "" {
				t.Errorf("reason %q with no member id", reason)
			}
		})
	}

	// Malformed events are ignored, never a panic.
	c := banSyncConfig()
	if id, _ := c.banSyncReasonForBan(&discordgo.GuildBanAdd{GuildID: syncGuild}); id != "" {
		t.Error("a ban with no user acted on")
	}
	if id, _ := c.banSyncReasonForUpdate(&discordgo.GuildMemberUpdate{}); id != "" {
		t.Error("an update with no member acted on")
	}
	if id, _ := c.banSyncReasonForRemove(nil); id != "" {
		t.Error("a nil remove acted on")
	}
}

func TestBanSyncIntents(t *testing.T) {
	members := discordgo.IntentGuildMembers
	moderation := discordgo.IntentGuildModeration

	base := Config{GuildIDs: []string{syncGuild}, PlayerRoleID: syncRole, CommunityGuildID: syncGuild}
	if got := base.Intents(); got&(members|moderation) != 0 {
		t.Errorf("ban sync off asks for %b; Server Members must never be asked for by default", got)
	}
	if got := base.Intents(); got&discordgo.IntentsGuilds == 0 {
		t.Error("the Guilds intent is missing")
	}

	on := base
	on.BanSync = true
	if got := on.Intents(); got&members == 0 || got&moderation == 0 {
		t.Errorf("ban sync on asks for %b, want Server Members and Guild Moderation", got)
	}

	// The flag without its guild or role is not on (Validate refuses it,
	// and nothing privileged is asked for meanwhile).
	half := on
	half.CommunityGuildID = ""
	if half.BanSyncEnabled() || half.Intents()&members != 0 {
		t.Error("ban sync with no community guild asks for Server Members")
	}
	half = on
	half.PlayerRoleID = ""
	if half.BanSyncEnabled() || half.Intents()&members != 0 {
		t.Error("ban sync with no player role asks for Server Members")
	}

	// Other features keep their own intents.
	cards := base
	cards.CardChannelIDs = []string{"1"}
	if got := cards.Intents(); got&discordgo.IntentMessageContent == 0 || got&members != 0 {
		t.Errorf("card watcher intents = %b", got)
	}
}

func TestBanSyncConfigValidate(t *testing.T) {
	good := Config{BotToken: "b", AppID: "a", GuildIDs: []string{"g"}, AdminToken: "t", ServerBaseURL: "http://x", ClientBaseURL: "https://x"}
	for _, tc := range []struct {
		name   string
		mutate func(*Config)
		wantIn string
	}{
		{"on without the guild", func(c *Config) { c.BanSync = true; c.PlayerRoleID = syncRole }, CommunityGuildEnv},
		{"on without the role", func(c *Config) { c.BanSync = true; c.CommunityGuildID = syncGuild }, PlayerRoleEnv},
		{"a malformed guild", func(c *Config) { c.CommunityGuildID = "c2" }, CommunityGuildEnv},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := good
			tc.mutate(&c)
			err := c.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.wantIn) {
				t.Fatalf("Validate = %v, want an error naming %s", err, tc.wantIn)
			}
		})
	}
	ok := good
	ok.BanSync, ok.CommunityGuildID, ok.PlayerRoleID = true, syncGuild, syncRole
	if err := ok.Validate(); err != nil {
		t.Errorf("a full ban sync config: %v", err)
	}
	// The guild alone is fine: the bot ignores it until ban sync is on.
	alone := good
	alone.CommunityGuildID = syncGuild
	if err := alone.Validate(); err != nil {
		t.Errorf("the community guild alone: %v", err)
	}
}

func TestBanSyncFromEnv(t *testing.T) {
	t.Setenv(BanSyncEnv, "1")
	t.Setenv(CommunityGuildEnv, " "+syncGuild+" ")
	c := ConfigFromEnv()
	if !c.BanSync || c.CommunityGuildID != syncGuild {
		t.Errorf("BanSync %v, CommunityGuildID %q", c.BanSync, c.CommunityGuildID)
	}
	for _, v := range []string{"", "0", "true", "yes"} {
		t.Setenv(BanSyncEnv, v)
		if ConfigFromEnv().BanSync {
			t.Errorf("%s=%q turned ban sync on; only 1 does", BanSyncEnv, v)
		}
	}
}

func TestOpenGatewayDropsServerMembersOn4014(t *testing.T) {
	refused := &websocket.CloseError{Code: 4014, Text: "Disallowed intent(s)."}
	withMembers := discordgo.IntentsGuilds | discordgo.IntentGuildModeration | discordgo.IntentGuildMembers

	t.Run("refused once, then open without it", func(t *testing.T) {
		intents := withMembers
		var seen []discordgo.Intent
		var logs bytes.Buffer
		err := openWithIntentFallback(func() error {
			seen = append(seen, intents)
			if intents&discordgo.IntentGuildMembers != 0 {
				return refused
			}
			return nil
		}, &intents, slog.New(slog.NewTextHandler(&logs, nil)))
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		if len(seen) != 2 {
			t.Fatalf("opened %d times, want 2", len(seen))
		}
		if intents&discordgo.IntentGuildMembers != 0 || intents&discordgo.IntentGuildModeration == 0 {
			t.Errorf("intents after the fallback = %b: Server Members must go and bans stay", intents)
		}
		if !strings.Contains(logs.String(), "level=ERROR") || !strings.Contains(logs.String(), "Server Members Intent") ||
			!strings.Contains(logs.String(), "Developer Portal") {
			t.Errorf("the error line does not say what to enable: %s", logs.String())
		}
	})

	t.Run("a second 4014 is returned", func(t *testing.T) {
		intents := withMembers | discordgo.IntentMessageContent
		calls := 0
		err := openWithIntentFallback(func() error { calls++; return refused }, &intents, discardLogger())
		if !IsDisallowedIntents(err) || calls != 2 {
			t.Errorf("err %v after %d opens, want the 4014 after 2", err, calls)
		}
	})

	t.Run("no Server Members asked for: no retry", func(t *testing.T) {
		intents := discordgo.IntentsGuilds | discordgo.IntentMessageContent
		calls := 0
		err := openWithIntentFallback(func() error { calls++; return refused }, &intents, discardLogger())
		if !IsDisallowedIntents(err) || calls != 1 {
			t.Errorf("err %v after %d opens, want the 4014 after 1", err, calls)
		}
	})

	t.Run("another failure: no retry", func(t *testing.T) {
		intents := withMembers
		calls := 0
		other := &websocket.CloseError{Code: 4004, Text: "Authentication failed."}
		err := openWithIntentFallback(func() error { calls++; return other }, &intents, discardLogger())
		if !errors.Is(err, other) || calls != 1 || intents != withMembers {
			t.Errorf("err %v after %d opens, intents %b", err, calls, intents)
		}
	})
}

func TestIsDisallowedIntents(t *testing.T) {
	if !IsDisallowedIntents(&websocket.CloseError{Code: 4014}) {
		t.Error("4014 not recognised")
	}
	if IsDisallowedIntents(&websocket.CloseError{Code: 4013}) || IsDisallowedIntents(errors.New("close 4014")) || IsDisallowedIntents(nil) {
		t.Error("something other than a 4014 close recognised")
	}
}

// banSyncServer is the game server's half: admin login and the ban sync
// route, answering with a canned outcome.
type banSyncServer struct {
	mu      sync.Mutex
	calls   []banSyncCall
	outcome string
	status  int
}

type banSyncCall struct{ path, reason, contentType string }

func newBanSyncServer(t *testing.T, outcome string) (*banSyncServer, *httptest.Server) {
	t.Helper()
	f := &banSyncServer{outcome: outcome, status: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /admin/login", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"token": "bot-session"})
	})
	mux.HandleFunc("POST /admin/discord-members/{discord_id}/revoke-sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer bot-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		var body struct {
			Reason string `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.calls = append(f.calls, banSyncCall{r.URL.Path, body.Reason, r.Header.Get("Content-Type")})
		status, outcome := f.status, f.outcome
		f.mu.Unlock()
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"outcome": outcome, "sockets_closed": 1})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return f, srv
}

func (f *banSyncServer) got() []banSyncCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]banSyncCall(nil), f.calls...)
}

func newBanSyncHandler(base string, logs *bytes.Buffer) (*Handler, *[]time.Duration) {
	h := NewHandler(banSyncConfig(), NewServerClient(base, "admin"), slog.New(slog.NewTextHandler(logs, nil)))
	var waits []time.Duration
	h.sleep = func(d time.Duration) { waits = append(waits, d) }
	return h, &waits
}

func TestBanSyncCallsTheServer(t *testing.T) {
	f, srv := newBanSyncServer(t, BanSyncRevoked)
	var logs bytes.Buffer
	h, waits := newBanSyncHandler(srv.URL, &logs)

	h.HandleBan(&discordgo.GuildBanAdd{GuildID: syncGuild, User: &discordgo.User{ID: memberID}})
	h.HandleMemberUpdate(&discordgo.GuildMemberUpdate{Member: guildMember(syncGuild, memberID), BeforeUpdate: guildMember(syncGuild, memberID, syncRole)})
	h.HandleMemberRemove(&discordgo.GuildMemberRemove{Member: guildMember(syncGuild, memberID)})
	// Not ban sync's: nothing is sent.
	h.HandleBan(&discordgo.GuildBanAdd{GuildID: otherGuild, User: &discordgo.User{ID: memberID}})
	h.HandleMemberUpdate(&discordgo.GuildMemberUpdate{Member: guildMember(syncGuild, memberID, syncRole)})

	got := f.got()
	want := []string{banSyncBanned, banSyncRoleRemoved, banSyncLeft}
	if len(got) != len(want) {
		t.Fatalf("calls = %+v, want %v", got, want)
	}
	for i, c := range got {
		if c.reason != want[i] || c.path != "/admin/discord-members/"+memberID+"/revoke-sessions" || c.contentType != "application/json" {
			t.Errorf("call %d = %+v, want reason %q", i, c, want[i])
		}
	}
	// The remove waited for a ban's own event first.
	if len(*waits) != 1 || (*waits)[0] != memberRemoveDelay {
		t.Errorf("waits = %v, want one of %v", *waits, memberRemoveDelay)
	}
	if !strings.Contains(logs.String(), "ban sync: sessions revoked") {
		t.Errorf("a revoke was not logged: %s", logs.String())
	}
}

// A Discord user with no account, an exempt one or one already out: the
// call is made, and nothing is logged.
func TestBanSyncIsQuietWhenNothingIsRevoked(t *testing.T) {
	for _, outcome := range []string{BanSyncNoAccount, BanSyncExempt, BanSyncAlreadyOut} {
		t.Run(outcome, func(t *testing.T) {
			f, srv := newBanSyncServer(t, outcome)
			var logs bytes.Buffer
			h, _ := newBanSyncHandler(srv.URL, &logs)
			h.HandleBan(&discordgo.GuildBanAdd{GuildID: syncGuild, User: &discordgo.User{ID: memberID}})
			if len(f.got()) != 1 {
				t.Fatalf("calls = %+v", f.got())
			}
			if logs.Len() != 0 {
				t.Errorf("logged for %s: %s", outcome, logs.String())
			}
		})
	}
}

func TestBanSyncRetriesAnUnreachableServerThenGivesUp(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close() // nothing listens: every call is ErrServerUnreachable
	var logs bytes.Buffer
	h, waits := newBanSyncHandler(base, &logs)
	h.HandleBan(&discordgo.GuildBanAdd{GuildID: syncGuild, User: &discordgo.User{ID: memberID}})
	if len(*waits) != banSyncAttempts-1 {
		t.Errorf("waited %d times, want %d", len(*waits), banSyncAttempts-1)
	}
	if !strings.Contains(logs.String(), "next renewal") {
		t.Errorf("giving up was not logged: %s", logs.String())
	}
}

func TestBanSyncDoesNotRetryARefusal(t *testing.T) {
	f, srv := newBanSyncServer(t, "")
	f.status = http.StatusBadRequest
	var logs bytes.Buffer
	h, waits := newBanSyncHandler(srv.URL, &logs)
	h.HandleBan(&discordgo.GuildBanAdd{GuildID: syncGuild, User: &discordgo.User{ID: memberID}})
	if len(f.got()) != 1 || len(*waits) != 0 {
		t.Errorf("calls %d, waits %v: a 400 is not retried", len(f.got()), *waits)
	}
}
