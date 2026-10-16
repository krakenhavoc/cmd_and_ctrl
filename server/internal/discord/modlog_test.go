package discord

// Tests for modlog.go (ADR 0156's 2026-10-16 amendment, #3325), against
// the httptest.Server voice_test.go stands up for Discord's REST API.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

const testModLogChannel = "777777777777777777"

func newModLogStub(t *testing.T) (*voiceStub, ModLog) {
	t.Helper()
	stub, _ := newVoiceStub(t)
	return stub, ModLog{Token: testToken, ChannelID: testModLogChannel}
}

func sampleModLogEntry() ModLogEntry {
	return ModLogEntry{
		Title: "Player report",
		Link:  "https://cmd.example/#/admin/reports",
		Fields: []ModLogField{
			{Name: "Table", Value: "Friday pod"},
			{Name: "Reported", Value: "Gus"},
			{Name: "Reason", Value: "kept typing slurs\nin chat"},
		},
		Footer: "Filed from the table",
		At:     time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC),
	}
}

func TestModLogFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing bool
	}{
		{name: "unset is off", vals: map[string]string{BotTokenEnv: testToken}},
		{name: "channel and token is on",
			vals:        map[string]string{ModLogChannelEnv: " " + testModLogChannel + " ", BotTokenEnv: testToken},
			wantEnabled: true},
		{name: "no token is off but boots",
			vals:        map[string]string{ModLogChannelEnv: testModLogChannel},
			wantMissing: true},
		{name: "a malformed channel fails",
			vals:    map[string]string{ModLogChannelEnv: "#mod-log", BotTokenEnv: testToken},
			wantErr: ModLogChannelEnv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ModLogFromEnv(env(tc.vals))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if m.Enabled() != tc.wantEnabled || m.MissingToken() != tc.wantMissing {
				t.Errorf("Enabled %v MissingToken %v, want %v %v", m.Enabled(), m.MissingToken(), tc.wantEnabled, tc.wantMissing)
			}
			if tc.wantEnabled && m.ChannelID != testModLogChannel {
				t.Errorf("channel not trimmed: %q", m.ChannelID)
			}
		})
	}
}

func TestModLogOffNeverCallsDiscord(t *testing.T) {
	stub, _ := newModLogStub(t)
	off := ModLog{ChannelID: testModLogChannel} // no token
	if err := off.Post(context.Background(), nil, sampleModLogEntry()); !errors.Is(err, ErrModLogNotConfigured) {
		t.Errorf("Post err = %v", err)
	}
	if n := len(stub.seen()); n != 0 {
		t.Errorf("%d calls reached Discord with the mod log off", n)
	}
}

func TestModLogPostEmbedAndNoMentions(t *testing.T) {
	stub, m := newModLogStub(t)
	stub.answer("POST", "/channels/"+testModLogChannel+"/messages", http.StatusOK, `{"id":"900"}`)
	if err := m.Post(context.Background(), nil, sampleModLogEntry()); err != nil {
		t.Fatal(err)
	}
	c := stub.seen()[0]
	if c.Method != "POST" || c.Auth != "Bot "+testToken || c.Reason != modLogAuditReason {
		t.Errorf("request = %s, auth %q, reason %q", c.Method, c.Auth, c.Reason)
	}
	if _, has := c.Body["content"]; has {
		t.Errorf("content = %v, want none", c.Body["content"])
	}
	am := c.Body["allowed_mentions"].(map[string]any)
	if parse, ok := am["parse"].([]any); !ok || len(parse) != 0 || am["roles"] != nil || am["users"] != nil {
		t.Errorf("allowed_mentions = %v, want parse: [] and nothing else", am)
	}
	embed := c.Body["embeds"].([]any)[0].(map[string]any)
	if embed["title"] != "Player report" || embed["url"] != "https://cmd.example/#/admin/reports" ||
		embed["timestamp"] != "2026-10-16T09:00:00Z" {
		t.Errorf("embed = %v", embed)
	}
	if f := embed["footer"].(map[string]any); f["text"] != "Filed from the table" {
		t.Errorf("footer = %v", f)
	}
	fields := embed["fields"].([]any)
	if len(fields) != 3 {
		t.Fatalf("fields = %v", fields)
	}
	reason := fields[2].(map[string]any)
	if reason["value"] != "kept typing slurs\nin chat" || reason["inline"] != false {
		t.Errorf("reason field = %v", reason)
	}
	pretty, _ := json.MarshalIndent(c.Body, "", "  ")
	t.Logf("example post body:\n%s", pretty)
}

func TestModLogHostileTextShowsAsTyped(t *testing.T) {
	e := ModLogEntry{
		Title:  "Seat removed",
		Link:   "javascript:alert(1)",
		Fields: []ModLogField{{Name: "Seat", Value: "<@123456789> **boss**"}, {Name: "Empty", Value: "   "}},
		Actor:  "Ada\nLovelace",
	}
	m := e.Render()
	em := m.Embeds[0]
	if em.URL != "" {
		t.Errorf("url = %q, want a non-http link dropped", em.URL)
	}
	if len(em.Fields) != 1 || em.Fields[0].Value != `\<\@123456789\> \*\*boss\*\*` {
		t.Errorf("fields = %+v", em.Fields)
	}
	if em.Footer.Text != "By Ada Lovelace" {
		t.Errorf("footer = %q", em.Footer.Text)
	}
	if len(m.AllowedMentions.Parse) != 0 || m.AllowedMentions.Users != nil || m.AllowedMentions.Roles != nil {
		t.Errorf("allowed_mentions = %+v", m.AllowedMentions)
	}
}

func TestModLogErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusForbidden, `{"message":"Missing Access","code":50001}`, ErrModLogForbidden},
		{http.StatusNotFound, `{"message":"Unknown Channel","code":10003}`, ErrModLogChannelNotFound},
		{http.StatusUnauthorized, `{}`, ErrBotUnauthorized},
		{http.StatusBadGateway, `{}`, ErrVoiceUnavailable},
	}
	for _, tc := range cases {
		stub, m := newModLogStub(t)
		stub.answer("POST", "/channels/"+testModLogChannel+"/messages", tc.status, tc.body)
		if err := m.Post(context.Background(), nil, sampleModLogEntry()); !errors.Is(err, tc.want) {
			t.Errorf("status %d: err = %v, want %v", tc.status, err, tc.want)
		}
	}
}
