package discord

// dm.go is the bot half of this package: opening a direct-message
// channel with a Discord user and posting one message into it, with
// the same stdlib net/http client oauth.go uses (ADR 0051 decision 5,
// S34 sub-PR 6).
//
// This is deliberately NOT the gateway bot. The bot binary
// (cmd/cmd_and_ctrl-bot) holds a websocket session for slash
// commands; this is two plain REST calls made by the SERVER process,
// so there is exactly one place that builds and sends an invite DM.
// The `/cc-invite-dm` slash command (#613) is a client of the route
// that calls this, not a second sender.
//
// The bot token never leaves this file's Authorization header: it is
// not logged, not wrapped into an error, and not echoed to a caller.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// botAPIBase is Discord's REST root for the two bot calls below.
// A package-level var, like oauth.go's endpoints, so tests can point
// it at an httptest.Server.
var botAPIBase = "https://discord.com/api/v10"

// BotTokenEnv is the environment variable the server reads the bot
// token from. Named here so every "not configured" message can say
// the same thing.
const BotTokenEnv = "CMDCTRL_DISCORD_BOT_TOKEN"

// maxDMBody caps how much of Discord's error body we read. The shapes
// we parse are a few hundred bytes.
const maxDMBody = 8 << 10

// Errors SendDM returns. Callers map them onto HTTP statuses; the
// text is safe to show a caller, and carries no token.
var (
	// ErrBotNotConfigured means no bot token is set on this
	// deployment. The route reports it as 503; it never falls open.
	ErrBotNotConfigured = errors.New("discord: " + BotTokenEnv + " is not set")
	// ErrDMForbidden is Discord's 403: the recipient shares no guild
	// with the bot, or has direct messages from server members
	// switched off. Indistinguishable from each other over the API.
	ErrDMForbidden = errors.New("discord: cannot open a DM with that user — they share no server with the bot, or have direct messages closed")
	// ErrDMRateLimited is Discord's 429 against the bot token. It is
	// the bot's own bucket, shared by every caller of this server, so
	// it is worth surfacing separately from our per-caller limiter.
	ErrDMRateLimited = errors.New("discord: rate-limited by Discord; try again shortly")
	// ErrBotUnauthorized is Discord's 401: the token is wrong or has
	// been reset in the Developer Portal.
	ErrBotUnauthorized = errors.New("discord: the bot token was rejected — it may have been reset")
	// ErrUnknownRecipient is Discord's 404 on the DM open: no such
	// user id.
	ErrUnknownRecipient = errors.New("discord: no Discord user with that id")
)

// Bot is the server's bot-token credential. The zero value is
// "unconfigured", and every method on it refuses rather than trying.
type Bot struct {
	// Token is the raw bot token (CMDCTRL_DISCORD_BOT_TOKEN). It is
	// sent as "Authorization: Bot <token>" and nowhere else.
	Token string
}

// BotFromEnv reads the bot token from the environment. A blank value
// yields a Bot whose Enabled() is false, which is a supported state:
// the DM-invite route answers 503 and every other route is unchanged.
func BotFromEnv(getenv func(string) string) Bot {
	return Bot{Token: strings.TrimSpace(getenv(BotTokenEnv))}
}

// Enabled reports whether a token is configured.
func (b Bot) Enabled() bool { return b.Token != "" }

// dmChannel is the subset of Discord's DM-channel object we need.
type dmChannel struct {
	ID string `json:"id"`
}

// apiError is the one field we read out of Discord's error body:
// `retry_after`, seconds as a float, on a 429. The body also carries
// `message` and `code`, and we deliberately do not forward either —
// see statusError.
type apiError struct {
	RetryAfter float64 `json:"retry_after"`
}

// SendDM opens a DM channel with recipientID and posts content into
// it: POST /users/@me/channels then POST /channels/{id}/messages,
// exactly as decision 5 describes.
//
// client may be nil, in which case a client with a short timeout is
// used. recipientID is a Discord snowflake.
func (b Bot) SendDM(ctx context.Context, client *http.Client, recipientID, content string) error {
	if !b.Enabled() {
		return ErrBotNotConfigured
	}
	if strings.TrimSpace(recipientID) == "" {
		return errors.New("discord: empty recipient id")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	var ch dmChannel
	if err := b.post(ctx, client, botAPIBase+"/users/@me/channels",
		map[string]string{"recipient_id": recipientID}, &ch); err != nil {
		return err
	}
	if ch.ID == "" {
		return errors.New("discord: DM channel response carried no id")
	}
	// allowed_mentions with an empty parse list makes the message
	// incapable of pinging anyone, whatever a display name interpolated
	// into it happens to contain.
	return b.post(ctx, client, botAPIBase+"/channels/"+ch.ID+"/messages",
		map[string]any{
			"content":          content,
			"allowed_mentions": map[string]any{"parse": []string{}},
		}, nil)
}

// post sends one authenticated JSON request and decodes a 2xx body
// into out (which may be nil to discard it). Non-2xx statuses map to
// the sentinels above.
func (b Bot) post(ctx context.Context, client *http.Client, url string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("discord: encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("discord: build request: %w", err)
	}
	// The only place the token appears. Never logged, never wrapped
	// into a returned error.
	req.Header.Set("Authorization", "Bot "+b.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// A transport error's text can contain the URL but never a
		// header, so this is safe to wrap.
		return fmt.Errorf("discord: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxDMBody))
	if err != nil {
		return fmt.Errorf("discord: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return statusError(resp.StatusCode, payload)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("discord: parse response: %w", err)
	}
	return nil
}

// statusError maps one non-2xx response onto a sentinel. The body is
// parsed only for the 429's retry_after; Discord's `message` is not
// forwarded verbatim, so nothing upstream can shape our error text.
func statusError(status int, body []byte) error {
	var ae apiError
	_ = json.Unmarshal(body, &ae)
	switch status {
	case http.StatusUnauthorized:
		return ErrBotUnauthorized
	case http.StatusForbidden:
		return ErrDMForbidden
	case http.StatusNotFound:
		return ErrUnknownRecipient
	case http.StatusTooManyRequests:
		if ae.RetryAfter > 0 {
			return fmt.Errorf("%w (retry after %.0fs)", ErrDMRateLimited, ae.RetryAfter)
		}
		return ErrDMRateLimited
	default:
		return fmt.Errorf("discord: DM failed: status %d", status)
	}
}
