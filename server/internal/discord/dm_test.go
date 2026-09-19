package discord

// Tests for dm.go — ADR 0051 decision 5's bot-token DM path (S34
// sub-PR 6). Everything here runs against an httptest.Server standing
// in for Discord's REST API; nothing reaches the network.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// dmAPIStub records what the bot sent and answers with whatever
// statuses the test asked for.
type dmAPIStub struct {
	mu sync.Mutex

	openStatus, msgStatus int
	openBody, msgBody     string

	openCalls, msgCalls int
	recipient           string
	content             string
	allowedMentions     any
	authHeaders         []string
	channelPath         string
}

func newDMAPIStub(t *testing.T) *dmAPIStub {
	t.Helper()
	stub := &dmAPIStub{
		openStatus: http.StatusOK,
		openBody:   `{"id":"dm-channel-1"}`,
		msgStatus:  http.StatusOK,
		msgBody:    `{"id":"msg-1"}`,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/users/@me/channels", func(w http.ResponseWriter, r *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.openCalls++
		stub.authHeaders = append(stub.authHeaders, r.Header.Get("Authorization"))
		var body struct {
			RecipientID string `json:"recipient_id"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		stub.recipient = body.RecipientID
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.openStatus)
		_, _ = w.Write([]byte(stub.openBody))
	})
	mux.HandleFunc("/channels/", func(w http.ResponseWriter, r *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.msgCalls++
		stub.channelPath = r.URL.Path
		stub.authHeaders = append(stub.authHeaders, r.Header.Get("Authorization"))
		var body struct {
			Content         string `json:"content"`
			AllowedMentions any    `json:"allowed_mentions"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		stub.content = body.Content
		stub.allowedMentions = body.AllowedMentions
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.msgStatus)
		_, _ = w.Write([]byte(stub.msgBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	t.Cleanup(SwapBotAPIBaseForTesting(srv.URL))
	return stub
}

func TestSendDMOpensAChannelThenPosts(t *testing.T) {
	stub := newDMAPIStub(t)
	b := Bot{Token: "bot-token-shhh"}

	if err := b.SendDM(context.Background(), nil, "snowflake-7", "come play"); err != nil {
		t.Fatalf("SendDM: %v", err)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if stub.openCalls != 1 || stub.msgCalls != 1 {
		t.Fatalf("calls: open %d, message %d; want 1 each", stub.openCalls, stub.msgCalls)
	}
	if stub.recipient != "snowflake-7" {
		t.Errorf("recipient_id = %q", stub.recipient)
	}
	if stub.channelPath != "/channels/dm-channel-1/messages" {
		t.Errorf("message posted to %q, want the channel the open returned", stub.channelPath)
	}
	if stub.content != "come play" {
		t.Errorf("content = %q", stub.content)
	}
	// A display name in the message must not be able to ping anyone.
	if stub.allowedMentions == nil {
		t.Error("message carried no allowed_mentions; a name in it could ping")
	}
	for _, h := range stub.authHeaders {
		if h != "Bot bot-token-shhh" {
			t.Errorf("Authorization = %q, want the bot scheme", h)
		}
	}
}

func TestSendDMWithoutATokenNeverCallsDiscord(t *testing.T) {
	stub := newDMAPIStub(t)
	var b Bot // the unconfigured zero value

	err := b.SendDM(context.Background(), nil, "snowflake-7", "hi")
	if err == nil || !strings.Contains(err.Error(), BotTokenEnv) {
		t.Fatalf("SendDM = %v, want an error naming %s", err, BotTokenEnv)
	}
	if !strings.Contains(err.Error(), ErrBotNotConfigured.Error()) {
		t.Errorf("SendDM = %v, want ErrBotNotConfigured", err)
	}
	stub.mu.Lock()
	defer stub.mu.Unlock()
	if stub.openCalls != 0 {
		t.Errorf("an unconfigured bot reached Discord %d times; it must never fall open", stub.openCalls)
	}
}

func TestSendDMMapsDiscordStatuses(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
		// onMessage puts the failure on the SECOND call instead of the
		// first, so both legs are covered.
		onMessage bool
	}{
		{name: "403 on open", status: http.StatusForbidden, body: `{"message":"Cannot send messages to this user"}`, want: ErrDMForbidden},
		{name: "403 on message", status: http.StatusForbidden, body: `{}`, want: ErrDMForbidden, onMessage: true},
		{name: "429", status: http.StatusTooManyRequests, body: `{"retry_after":4.2}`, want: ErrDMRateLimited},
		{name: "401", status: http.StatusUnauthorized, body: `{}`, want: ErrBotUnauthorized},
		{name: "404", status: http.StatusNotFound, body: `{}`, want: ErrUnknownRecipient},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := newDMAPIStub(t)
			stub.mu.Lock()
			if tc.onMessage {
				stub.msgStatus, stub.msgBody = tc.status, tc.body
			} else {
				stub.openStatus, stub.openBody = tc.status, tc.body
			}
			stub.mu.Unlock()

			err := Bot{Token: "bot-token-shhh"}.SendDM(context.Background(), nil, "snowflake-7", "hi")
			if err == nil {
				t.Fatal("SendDM succeeded against a failing Discord")
			}
			if !strings.Contains(err.Error(), tc.want.Error()) {
				t.Fatalf("SendDM = %v, want %v", err, tc.want)
			}
			// Never the token, and never Discord's own message text.
			if strings.Contains(err.Error(), "bot-token-shhh") {
				t.Fatal("the bot token leaked into an error")
			}
			if strings.Contains(err.Error(), "Cannot send messages to this user") {
				t.Errorf("Discord's error text is forwarded verbatim: %v", err)
			}
			if tc.status == http.StatusTooManyRequests && !strings.Contains(err.Error(), "4s") {
				t.Errorf("a 429 should surface retry_after: %v", err)
			}
		})
	}
}

func TestBotFromEnvTrimsAndReportsEnabled(t *testing.T) {
	env := map[string]string{BotTokenEnv: "  tok  "}
	b := BotFromEnv(func(k string) string { return env[k] })
	if b.Token != "tok" || !b.Enabled() {
		t.Errorf("BotFromEnv = %+v, enabled %v", b, b.Enabled())
	}
	blank := BotFromEnv(func(string) string { return "   " })
	if blank.Enabled() {
		t.Error("a whitespace-only token reports itself configured")
	}
}
