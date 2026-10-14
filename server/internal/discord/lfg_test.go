package discord

// Tests for lfg.go (ADR 0161 §2, §3, §5). Everything runs against the
// httptest.Server voice_test.go stands up for Discord's REST API.

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	testChannel = "444444444444444444"
	testRole    = "555555555555555555"
)

func newLFGStub(t *testing.T) (*voiceStub, LFG) {
	t.Helper()
	stub, _ := newVoiceStub(t)
	return stub, LFG{Token: testToken, GuildID: testGuild, ChannelID: testChannel, RoleID: testRole}
}

func samplePost() LFGPost {
	return LFGPost{
		Table:     "  Friday pod  ",
		HostID:    "501",
		Seated:    2,
		Seats:     4,
		Voice:     true,
		Settings:  []LFGField{{Name: "Starting life", Value: "30"}},
		CreatedAt: time.Date(2026, 10, 13, 18, 0, 0, 0, time.UTC),
		JoinURL:   "https://cmdctrl.example/#/games/g1/join?t=tok",
	}
}

func TestLFGFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing bool
		wantRole    string
	}{
		{name: "unset is off", vals: map[string]string{}},
		{name: "guild and token without a channel is off",
			vals: map[string]string{CommunityGuildEnv: testGuild, BotTokenEnv: testToken}},
		{name: "channel, guild and token is on",
			vals:        map[string]string{LFGChannelEnv: " " + testChannel + " ", CommunityGuildEnv: testGuild, BotTokenEnv: testToken},
			wantEnabled: true},
		{name: "the role is optional and read when set",
			vals:        map[string]string{LFGChannelEnv: testChannel, LFGRoleEnv: testRole, CommunityGuildEnv: testGuild, BotTokenEnv: testToken},
			wantEnabled: true, wantRole: testRole},
		{name: "no token is off but boots",
			vals:        map[string]string{LFGChannelEnv: testChannel, LFGRoleEnv: testRole, CommunityGuildEnv: testGuild},
			wantMissing: true, wantRole: testRole},
		{name: "a role without the channel fails",
			vals:    map[string]string{LFGRoleEnv: testRole, CommunityGuildEnv: testGuild, BotTokenEnv: testToken},
			wantErr: LFGChannelEnv},
		{name: "the channel without the guild fails",
			vals:    map[string]string{LFGChannelEnv: testChannel, BotTokenEnv: testToken},
			wantErr: CommunityGuildEnv},
		{name: "a malformed channel fails",
			vals:    map[string]string{LFGChannelEnv: "#lfg", CommunityGuildEnv: testGuild},
			wantErr: LFGChannelEnv},
		{name: "a malformed role fails",
			vals:    map[string]string{LFGChannelEnv: testChannel, LFGRoleEnv: "@LFG", CommunityGuildEnv: testGuild},
			wantErr: LFGRoleEnv},
		{name: "a malformed guild fails",
			vals:    map[string]string{LFGChannelEnv: testChannel, CommunityGuildEnv: "c2"},
			wantErr: CommunityGuildEnv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := LFGFromEnv(env(tc.vals))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if f.Enabled() != tc.wantEnabled || f.MissingToken() != tc.wantMissing {
				t.Errorf("Enabled %v MissingToken %v, want %v %v", f.Enabled(), f.MissingToken(), tc.wantEnabled, tc.wantMissing)
			}
			if f.RoleID != tc.wantRole {
				t.Errorf("RoleID = %q, want %q", f.RoleID, tc.wantRole)
			}
			if tc.wantEnabled && f.ChannelID != testChannel {
				t.Errorf("channel not trimmed: %q", f.ChannelID)
			}
		})
	}
}

func TestLFGOffNeverCallsDiscord(t *testing.T) {
	stub, _ := newLFGStub(t)
	off := LFG{GuildID: testGuild, ChannelID: testChannel} // no token
	ctx := context.Background()
	if _, err := off.Post(ctx, nil, samplePost(), true); !errors.Is(err, ErrLFGNotConfigured) {
		t.Errorf("Post err = %v", err)
	}
	if err := off.Edit(ctx, nil, "1", samplePost()); !errors.Is(err, ErrLFGNotConfigured) {
		t.Errorf("Edit err = %v", err)
	}
	if err := off.Delete(ctx, nil, "1"); !errors.Is(err, ErrLFGNotConfigured) {
		t.Errorf("Delete err = %v", err)
	}
	if n := len(stub.seen()); n != 0 {
		t.Errorf("%d calls reached Discord with the board off", n)
	}
}

func TestLFGPostPingsOnlyTheRole(t *testing.T) {
	stub, f := newLFGStub(t)
	stub.answer("POST", "/channels/"+testChannel+"/messages", http.StatusOK, `{"id":"900"}`)

	id, err := f.Post(context.Background(), nil, samplePost(), true)
	if err != nil || id != "900" {
		t.Fatalf("Post = %q, %v", id, err)
	}
	c := stub.seen()[0]
	if c.Method != "POST" || c.Auth != "Bot "+testToken {
		t.Errorf("request = %s, auth %q", c.Method, c.Auth)
	}
	if c.Body["content"] != "<@&"+testRole+">" {
		t.Errorf("content = %v, want the role mention", c.Body["content"])
	}
	am := c.Body["allowed_mentions"].(map[string]any)
	if parse, _ := am["parse"].([]any); len(parse) != 0 {
		t.Errorf("allowed_mentions.parse = %v, want empty", am["parse"])
	}
	if roles, _ := am["roles"].([]any); len(roles) != 1 || roles[0] != testRole {
		t.Errorf("allowed_mentions.roles = %v, want only the LFG role", am["roles"])
	}

	embed := c.Body["embeds"].([]any)[0].(map[string]any)
	if embed["title"] != "Friday pod" || embed["timestamp"] != "2026-10-13T18:00:00Z" {
		t.Errorf("embed = %v", embed)
	}
	fields := map[string]string{}
	for _, raw := range embed["fields"].([]any) {
		m := raw.(map[string]any)
		fields[m["name"].(string)] = m["value"].(string)
	}
	want := map[string]string{"Host": "<@501>", "Seats": "2 of 4", "Voice room": "Yes", "Starting life": "30"}
	for k, v := range want {
		if fields[k] != v {
			t.Errorf("field %s = %q, want %q", k, fields[k], v)
		}
	}
	row := c.Body["components"].([]any)[0].(map[string]any)
	btn := row["components"].([]any)[0].(map[string]any)
	if btn["style"] != float64(buttonStyleLink) || btn["url"] != samplePost().JoinURL ||
		btn["label"] != "Join table" || btn["disabled"] != nil {
		t.Errorf("button = %v", btn)
	}
}

func TestLFGPostWithoutPingOrRoleMentionsNobody(t *testing.T) {
	stub, f := newLFGStub(t)
	stub.answer("POST", "/channels/"+testChannel+"/messages", http.StatusOK, `{"id":"900"}`)
	ctx := context.Background()
	if _, err := f.Post(ctx, nil, samplePost(), false); err != nil {
		t.Fatal(err)
	}
	noRole := f
	noRole.RoleID = ""
	if _, err := noRole.Post(ctx, nil, samplePost(), true); err != nil {
		t.Fatal(err)
	}
	for i, c := range stub.seen() {
		am := c.Body["allowed_mentions"].(map[string]any)
		if c.Body["content"] != "" || am["roles"] != nil {
			t.Errorf("post %d: content %v, allowed roles %v; want no ping", i, c.Body["content"], am["roles"])
		}
	}
}

func TestLFGPostNeedsTheJoinURL(t *testing.T) {
	stub, f := newLFGStub(t)
	p := samplePost()
	p.JoinURL = ""
	if _, err := f.Post(context.Background(), nil, p, true); err == nil {
		t.Fatal("Post without a join link succeeded")
	}
	if len(stub.seen()) != 0 {
		t.Error("Post without a join link reached Discord")
	}
}

func TestLFGEditNeverPingsAndDisablesAFullTable(t *testing.T) {
	stub, f := newLFGStub(t)
	p := samplePost()
	p.Seated = 4
	if err := f.Edit(context.Background(), nil, "900", p); err != nil {
		t.Fatal(err)
	}
	c := stub.seen()[0]
	if c.Method != "PATCH" || c.Path != "/channels/"+testChannel+"/messages/900" {
		t.Errorf("edit = %s %s", c.Method, c.Path)
	}
	if _, ok := c.Body["content"]; ok {
		t.Errorf("an edit sent content %v", c.Body["content"])
	}
	am := c.Body["allowed_mentions"].(map[string]any)
	if am["roles"] != nil {
		t.Errorf("an edit allowed a role mention: %v", am)
	}
	btn := c.Body["components"].([]any)[0].(map[string]any)["components"].([]any)[0].(map[string]any)
	if btn["disabled"] != true || btn["label"] != "Table full" {
		t.Errorf("full table's button = %v", btn)
	}
	embed := c.Body["embeds"].([]any)[0].(map[string]any)
	seats := embed["fields"].([]any)[1].(map[string]any)
	if seats["value"] != "4 of 4 · Full" {
		t.Errorf("seats = %v", seats)
	}
}

func TestLFGEditWithoutTheLinkLeavesTheButton(t *testing.T) {
	stub, f := newLFGStub(t)
	p := samplePost()
	p.JoinURL = ""
	if err := f.Edit(context.Background(), nil, "900", p); err != nil {
		t.Fatal(err)
	}
	if _, ok := stub.seen()[0].Body["components"]; ok {
		t.Error("an edit with no known link touched the button")
	}
}

func TestLFGDeleteAndGoneMessages(t *testing.T) {
	stub, f := newLFGStub(t)
	ctx := context.Background()
	msg := "/channels/" + testChannel + "/messages/"
	if err := f.Delete(ctx, nil, "900"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if c := stub.seen()[0]; c.Method != "DELETE" || c.Path != msg+"900" || c.Reason != lfgAuditReason {
		t.Errorf("delete = %+v", c)
	}
	// Already gone: success for a delete, ErrLFGMessageGone for an edit.
	stub.answer("DELETE", msg+"901", http.StatusNotFound, `{"code":10008}`)
	if err := f.Delete(ctx, nil, "901"); err != nil {
		t.Errorf("Delete of a gone message: %v", err)
	}
	stub.answer("PATCH", msg+"901", http.StatusNotFound, `{"code":10008}`)
	if err := f.Edit(ctx, nil, "901", samplePost()); !errors.Is(err, ErrLFGMessageGone) {
		t.Errorf("Edit of a gone message: %v", err)
	}
	// A 404 for the channel is not "gone": the configuration is wrong.
	stub.answer("DELETE", msg+"902", http.StatusNotFound, `{"code":10003}`)
	if err := f.Delete(ctx, nil, "902"); !errors.Is(err, ErrLFGChannelNotFound) {
		t.Errorf("Delete in an unknown channel: %v", err)
	}
	stub.answer("POST", "/channels/"+testChannel+"/messages", http.StatusNotFound, `{"code":10003}`)
	if _, err := f.Post(ctx, nil, samplePost(), false); !errors.Is(err, ErrLFGChannelNotFound) || Transient(err) {
		t.Errorf("Post in an unknown channel: %v", err)
	}
}

func TestLFGStatusMapping(t *testing.T) {
	stub, f := newLFGStub(t)
	ctx := context.Background()
	path := "/channels/" + testChannel + "/messages/900"
	cases := []struct {
		status    int
		body      string
		want      error
		transient bool
		wait      time.Duration
	}{
		{http.StatusUnauthorized, `{"message":"401: Unauthorized ` + testToken + `"}`, ErrBotUnauthorized, false, 0},
		{http.StatusForbidden, `{"code":50013,"message":"Missing Permissions"}`, ErrLFGForbidden, false, 0},
		{http.StatusTooManyRequests, `{"retry_after":1.5}`, ErrVoiceRateLimited, true, 1500 * time.Millisecond},
		{http.StatusServiceUnavailable, `upstream`, ErrVoiceUnavailable, true, 0},
	}
	for _, tc := range cases {
		stub.answer("PATCH", path, tc.status, tc.body)
		err := f.Edit(ctx, nil, "900", samplePost())
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: err = %v, want %v", tc.status, err, tc.want)
		}
		if Transient(err) != tc.transient {
			t.Errorf("status %d: Transient = %v", tc.status, Transient(err))
		}
		if tc.wait > 0 {
			var rl *RateLimitError
			if !errors.As(err, &rl) || rl.RetryAfter != tc.wait {
				t.Errorf("status %d: retry after = %+v, want %s", tc.status, rl, tc.wait)
			}
		}
		if strings.Contains(err.Error(), testToken) || strings.Contains(err.Error(), "Missing Permissions") {
			t.Errorf("status %d: error forwards Discord's text or the token: %v", tc.status, err)
		}
	}
}

func TestLFGUnreachableIsTransient(t *testing.T) {
	f := LFG{Token: testToken, GuildID: testGuild, ChannelID: testChannel}
	restore := SwapBotAPIBaseForTesting("http://127.0.0.1:1")
	defer restore()
	err := f.Delete(context.Background(), nil, "900")
	if !errors.Is(err, ErrVoiceUnavailable) || !Transient(err) {
		t.Errorf("err = %v, want a transient ErrVoiceUnavailable", err)
	}
	if strings.Contains(err.Error(), testToken) {
		t.Error("transport error carries the token")
	}
}

func TestPostTitle(t *testing.T) {
	long := strings.Repeat("é", 300)
	cases := map[string]string{
		"  Friday pod ": "Friday pod",
		"":              "Table",
		long:            strings.Repeat("é", 256),
	}
	for in, want := range cases {
		if got := PostTitle(in); got != want {
			t.Errorf("PostTitle(%q) = %q, want %q", in, got, want)
		}
	}
}
