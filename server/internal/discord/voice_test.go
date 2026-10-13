package discord

// Tests for voice.go (ADR 0159 §4). Everything runs against an
// httptest.Server standing in for Discord's REST API.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testGuild    = "111111111111111111"
	testCategory = "222222222222222222"
	testBot      = "333333333333333333"
	testToken    = "voice-test-token"
)

// voiceCall is one request the stub saw.
type voiceCall struct {
	Method, Path, Auth, Reason string
	Body                       map[string]any
}

// voiceStub answers each "METHOD /path" with a canned status and body,
// and records every request.
type voiceStub struct {
	mu      sync.Mutex
	answers map[string]stubAnswer
	calls   []voiceCall
}

type stubAnswer struct {
	status int
	body   string
	header map[string]string
}

func newVoiceStub(t *testing.T) (*voiceStub, Voice) {
	t.Helper()
	stub := &voiceStub{answers: map[string]stubAnswer{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		stub.mu.Lock()
		stub.calls = append(stub.calls, voiceCall{
			Method: r.Method, Path: r.URL.Path,
			Auth: r.Header.Get("Authorization"), Reason: r.Header.Get("X-Audit-Log-Reason"),
			Body: body,
		})
		a, ok := stub.answers[r.Method+" "+r.URL.Path]
		stub.mu.Unlock()
		if !ok {
			a = stubAnswer{status: http.StatusNoContent}
		}
		for k, v := range a.header {
			w.Header().Set(k, v)
		}
		w.WriteHeader(a.status)
		_, _ = w.Write([]byte(a.body))
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(SwapBotAPIBaseForTesting(srv.URL))
	return stub, Voice{Token: testToken, GuildID: testGuild, CategoryID: testCategory}
}

func (s *voiceStub) answer(method, path string, status int, body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.answers[method+" "+path] = stubAnswer{status: status, body: body}
}

func (s *voiceStub) seen() []voiceCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]voiceCall(nil), s.calls...)
}

func env(vals map[string]string) func(string) string {
	return func(k string) string { return vals[k] }
}

func TestVoiceFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing bool
	}{
		{name: "unset is off", vals: map[string]string{}},
		{name: "guild and token without a category is off",
			vals: map[string]string{CommunityGuildEnv: testGuild, BotTokenEnv: testToken}},
		{name: "all three is on",
			vals:        map[string]string{VoiceCategoryEnv: testCategory, CommunityGuildEnv: testGuild, BotTokenEnv: " " + testToken + " "},
			wantEnabled: true},
		{name: "no token is off but boots",
			vals:        map[string]string{VoiceCategoryEnv: testCategory, CommunityGuildEnv: testGuild},
			wantMissing: true},
		{name: "category without a guild fails",
			vals:    map[string]string{VoiceCategoryEnv: testCategory, BotTokenEnv: testToken},
			wantErr: CommunityGuildEnv},
		{name: "malformed category fails",
			vals:    map[string]string{VoiceCategoryEnv: "table-voice", CommunityGuildEnv: testGuild},
			wantErr: VoiceCategoryEnv},
		{name: "malformed guild fails",
			vals:    map[string]string{VoiceCategoryEnv: testCategory, CommunityGuildEnv: "c2"},
			wantErr: CommunityGuildEnv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := VoiceFromEnv(env(tc.vals))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if v.Enabled() != tc.wantEnabled {
				t.Errorf("Enabled() = %v, want %v", v.Enabled(), tc.wantEnabled)
			}
			if v.MissingToken() != tc.wantMissing {
				t.Errorf("MissingToken() = %v, want %v", v.MissingToken(), tc.wantMissing)
			}
			if tc.wantEnabled && v.Token != testToken {
				t.Errorf("token not trimmed: %q", v.Token)
			}
		})
	}
}

func TestVoiceOffNeverCallsDiscord(t *testing.T) {
	stub, _ := newVoiceStub(t)
	off := Voice{GuildID: testGuild, CategoryID: testCategory} // no token
	ctx := context.Background()
	if _, err := off.CreateRoom(ctx, nil, "t", testBot, nil); !errors.Is(err, ErrVoiceNotConfigured) {
		t.Errorf("CreateRoom err = %v", err)
	}
	if _, err := off.IsMember(ctx, nil, "1"); !errors.Is(err, ErrVoiceNotConfigured) {
		t.Errorf("IsMember err = %v", err)
	}
	if err := off.DeleteRoom(ctx, nil, "1"); !errors.Is(err, ErrVoiceNotConfigured) {
		t.Errorf("DeleteRoom err = %v", err)
	}
	if n := len(stub.seen()); n != 0 {
		t.Errorf("%d calls reached Discord with voice off", n)
	}
}

func TestCreateRoomIsPrivateToTheBotAndTheMembers(t *testing.T) {
	stub, v := newVoiceStub(t)
	stub.answer("POST", "/guilds/"+testGuild+"/channels", http.StatusCreated, `{"id":"444"}`)

	id, err := v.CreateRoom(context.Background(), nil, "  Friday pod  ", testBot, []string{"501", "", "502", testBot})
	if err != nil || id != "444" {
		t.Fatalf("CreateRoom = %q, %v", id, err)
	}
	calls := stub.seen()
	if len(calls) != 1 {
		t.Fatalf("calls = %d, want 1", len(calls))
	}
	c := calls[0]
	if c.Auth != "Bot "+testToken || c.Reason != auditReason {
		t.Errorf("headers: auth %q reason %q", c.Auth, c.Reason)
	}
	if c.Body["name"] != "Friday pod" || c.Body["parent_id"] != testCategory || c.Body["type"] != float64(channelTypeVoice) {
		t.Errorf("body = %v", c.Body)
	}
	ows, _ := c.Body["permission_overwrites"].([]any)
	got := map[string]map[string]any{}
	for _, o := range ows {
		m := o.(map[string]any)
		got[m["id"].(string)] = m
	}
	if len(got) != 4 {
		t.Fatalf("overwrites = %v, want @everyone, the bot and two players (blank and bot ids skipped)", ows)
	}
	every := got[testGuild]
	if every["type"] != float64(overwriteRole) || every["deny"] != perms(EveryoneDeny) || every["allow"] != "0" {
		t.Errorf("@everyone overwrite = %v", every)
	}
	if got[testBot]["allow"] != perms(BotAllow) || got[testBot]["type"] != float64(overwriteMember) {
		t.Errorf("bot overwrite = %v", got[testBot])
	}
	for _, p := range []string{"501", "502"} {
		if got[p]["allow"] != perms(PlayerAllow) || got[p]["type"] != float64(overwriteMember) {
			t.Errorf("player %s overwrite = %v", p, got[p])
		}
	}
}

func TestCreateRoomNeedsTheBotID(t *testing.T) {
	stub, v := newVoiceStub(t)
	if _, err := v.CreateRoom(context.Background(), nil, "t", "", nil); err == nil {
		t.Fatal("CreateRoom without a bot id succeeded")
	}
	if len(stub.seen()) != 0 {
		t.Error("CreateRoom without a bot id reached Discord")
	}
}

func TestGrantAndRevokeMember(t *testing.T) {
	stub, v := newVoiceStub(t)
	ctx := context.Background()
	if err := v.GrantMember(ctx, nil, "444", "501"); err != nil {
		t.Fatalf("GrantMember: %v", err)
	}
	if err := v.RevokeMember(ctx, nil, "444", "501"); err != nil {
		t.Fatalf("RevokeMember: %v", err)
	}
	calls := stub.seen()
	if calls[0].Method != "PUT" || calls[0].Path != "/channels/444/permissions/501" ||
		calls[0].Body["allow"] != perms(PlayerAllow) || calls[0].Body["type"] != float64(overwriteMember) {
		t.Errorf("grant = %+v", calls[0])
	}
	if calls[1].Method != "DELETE" || calls[1].Path != "/channels/444/permissions/501" {
		t.Errorf("revoke = %+v", calls[1])
	}
}

func TestRemovingWhatIsAlreadyGoneSucceeds(t *testing.T) {
	stub, v := newVoiceStub(t)
	ctx := context.Background()
	stub.answer("DELETE", "/channels/444", http.StatusNotFound, `{"code":10003}`)
	stub.answer("DELETE", "/channels/444/permissions/501", http.StatusNotFound, `{"code":10009}`)
	if err := v.DeleteRoom(ctx, nil, "444"); err != nil {
		t.Errorf("DeleteRoom of a gone room: %v", err)
	}
	if err := v.RevokeMember(ctx, nil, "444", "501"); err != nil {
		t.Errorf("RevokeMember of a gone overwrite: %v", err)
	}
	// A 404 for something else (the guild) is still an error.
	stub.answer("DELETE", "/channels/555", http.StatusNotFound, `{"code":10004}`)
	if err := v.DeleteRoom(ctx, nil, "555"); !errors.Is(err, ErrVoiceNotFound) {
		t.Errorf("DeleteRoom with an unknown guild: %v", err)
	}
}

func TestRoomsListsOnlyVoiceChannelsInTheCategory(t *testing.T) {
	stub, v := newVoiceStub(t)
	stub.answer("GET", "/guilds/"+testGuild+"/channels", http.StatusOK, `[
		{"id":"1","name":"general","type":0,"parent_id":null},
		{"id":"2","name":"Friday pod","type":2,"parent_id":"`+testCategory+`"},
		{"id":"3","name":"General","type":2,"parent_id":"999"},
		{"id":"4","name":"notes","type":0,"parent_id":"`+testCategory+`"},
		{"id":"5","name":"Second pod","type":2,"parent_id":"`+testCategory+`"}
	]`)
	rooms, err := v.Rooms(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 2 || rooms[0] != (Room{ID: "2", Name: "Friday pod"}) || rooms[1].ID != "5" {
		t.Errorf("rooms = %+v", rooms)
	}
}

func TestIsMember(t *testing.T) {
	stub, v := newVoiceStub(t)
	ctx := context.Background()
	stub.answer("GET", "/guilds/"+testGuild+"/members/501", http.StatusOK, `{"user":{"id":"501"}}`)
	stub.answer("GET", "/guilds/"+testGuild+"/members/502", http.StatusNotFound, `{"code":10007}`)
	stub.answer("GET", "/guilds/"+testGuild+"/members/503", http.StatusNotFound, `{"code":10013}`)
	stub.answer("GET", "/guilds/"+testGuild+"/members/504", http.StatusNotFound, `{"code":10004}`)

	for id, want := range map[string]bool{"501": true, "502": false, "503": false} {
		got, err := v.IsMember(ctx, nil, id)
		if err != nil || got != want {
			t.Errorf("IsMember(%s) = %v, %v; want %v", id, got, err, want)
		}
	}
	if _, err := v.IsMember(ctx, nil, "504"); !errors.Is(err, ErrVoiceNotFound) {
		t.Errorf("IsMember with an unknown guild: err = %v", err)
	}
}

func TestDisconnectFromOnlyTouchesThisRoom(t *testing.T) {
	stub, v := newVoiceStub(t)
	ctx := context.Background()
	vs := "/guilds/" + testGuild + "/voice-states/"
	stub.answer("GET", vs+"501", http.StatusOK, `{"channel_id":"444"}`)
	stub.answer("GET", vs+"502", http.StatusOK, `{"channel_id":"999"}`)
	stub.answer("GET", vs+"503", http.StatusNotFound, `{"code":10065}`)
	stub.answer("PATCH", "/guilds/"+testGuild+"/members/501", http.StatusOK, `{}`)

	for id, want := range map[string]bool{"501": true, "502": false, "503": false} {
		got, err := v.DisconnectFrom(ctx, nil, "444", id)
		if err != nil || got != want {
			t.Errorf("DisconnectFrom(%s) = %v, %v; want %v", id, got, err, want)
		}
	}
	var patches []voiceCall
	for _, c := range stub.seen() {
		if c.Method == "PATCH" {
			patches = append(patches, c)
		}
	}
	if len(patches) != 1 || patches[0].Path != "/guilds/"+testGuild+"/members/501" {
		t.Fatalf("patches = %+v, want only player 501's", patches)
	}
	if v, ok := patches[0].Body["channel_id"]; !ok || v != nil {
		t.Errorf("disconnect body = %v, want channel_id: null", patches[0].Body)
	}
}

func TestBotUserID(t *testing.T) {
	stub, v := newVoiceStub(t)
	stub.answer("GET", "/users/@me", http.StatusOK, `{"id":"`+testBot+`"}`)
	id, err := v.BotUserID(context.Background(), nil)
	if err != nil || id != testBot {
		t.Fatalf("BotUserID = %q, %v", id, err)
	}
	if r := stub.seen()[0].Reason; r != "" {
		t.Errorf("a GET carried an audit reason %q", r)
	}
}

func TestVoiceStatusMapping(t *testing.T) {
	stub, v := newVoiceStub(t)
	ctx := context.Background()
	path := "/channels/444/permissions/501"
	cases := []struct {
		status    int
		body      string
		want      error
		transient bool
		wait      time.Duration
	}{
		{http.StatusUnauthorized, `{"message":"401: Unauthorized ` + testToken + `"}`, ErrBotUnauthorized, false, 0},
		{http.StatusForbidden, `{"code":50013,"message":"Missing Permissions"}`, ErrVoiceForbidden, false, 0},
		{http.StatusTooManyRequests, `{"retry_after":2.5}`, ErrVoiceRateLimited, true, 2500 * time.Millisecond},
		{http.StatusBadGateway, `upstream`, ErrVoiceUnavailable, true, 0},
	}
	for _, tc := range cases {
		stub.answer("PUT", path, tc.status, tc.body)
		err := v.GrantMember(ctx, nil, "444", "501")
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

func TestVoiceUnreachableIsTransient(t *testing.T) {
	v := Voice{Token: testToken, GuildID: testGuild, CategoryID: testCategory}
	restore := SwapBotAPIBaseForTesting("http://127.0.0.1:1")
	defer restore()
	err := v.DeleteRoom(context.Background(), nil, "444")
	if !errors.Is(err, ErrVoiceUnavailable) || !Transient(err) {
		t.Errorf("err = %v, want a transient ErrVoiceUnavailable", err)
	}
	if strings.Contains(err.Error(), testToken) {
		t.Error("transport error carries the token")
	}
}

func TestRoomName(t *testing.T) {
	long := strings.Repeat("é", 120)
	cases := map[string]string{
		"  Friday pod ": "Friday pod",
		"":              "Table",
		"   ":           "Table",
		long:            strings.Repeat("é", 100),
	}
	for in, want := range cases {
		if got := RoomName(in); got != want {
			t.Errorf("RoomName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRoomURL(t *testing.T) {
	v := Voice{GuildID: testGuild}
	if got := v.RoomURL("444"); got != "https://discord.com/channels/"+testGuild+"/444" {
		t.Errorf("RoomURL = %q", got)
	}
}
