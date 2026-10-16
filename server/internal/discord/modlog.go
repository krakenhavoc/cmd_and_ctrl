package discord

// modlog.go is ADR 0156's 2026-10-16 amendment (#3325): the REST call
// behind the mod log, a private channel in the community server where the
// server posts each in-app player report and each admin action (a table
// ended or deleted, a seat or spectator removed, an account's sessions
// revoked).
//
// Same rules as results.go and bugforum.go, whose helpers it reuses: the
// SERVER calls Discord with the bot token (not the gateway bot binary),
// the token appears in the Authorization header and nowhere else, every
// name and reason is escaped so it shows as typed, and nobody is
// mentioned.
//
// Nothing here decides WHAT is logged. That is the lobby (lobby/modlog.go);
// this file only says what an entry looks like on the wire and sends it.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ModLogChannelEnv names the private #mod-log channel. Unset: nothing is
// posted. The gateway bot reads the same variable for its own entries
// (an access request approved or declined).
const ModLogChannelEnv = "CMDCTRL_DISCORD_MOD_LOG_CHANNEL_ID"

// modLogAuditReason labels the posts in the server's audit log.
const modLogAuditReason = "CMD & CTRL mod log"

// modLogColor is the embed's side bar: Discord's yellow, so an entry
// reads differently from a bug thread's red and a result's green.
const modLogColor = 0xFEE75C

// maxModLogFields caps an entry's fields, well under Discord's 25.
const maxModLogFields = 8

// Errors the mod log call returns, beside voice.go's ErrVoiceRateLimited
// (a *RateLimitError), ErrVoiceUnavailable and ErrBotUnauthorized, which
// mean the same thing for a message. Text is safe to log.
var (
	// ErrModLogNotConfigured: the mod log is off on this deployment.
	ErrModLogNotConfigured = errors.New("discord: the mod log is off — " + ModLogChannelEnv + " and " + BotTokenEnv + " must both be set")
	// ErrModLogForbidden is Discord's 403: the bot cannot post in the
	// channel.
	ErrModLogForbidden = errors.New("discord: the bot cannot post in the mod log channel — check View Channel, Send Messages and Embed Links there")
	// ErrModLogChannelNotFound is a 404: the channel id is wrong, or the
	// bot is not in the server.
	ErrModLogChannelNotFound = errors.New("discord: the mod log channel was not found — is the bot in the server?")
)

// ModLog is the mod log's configuration. The zero value is "off", and
// every call on it refuses without touching the network.
type ModLog struct {
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN).
	Token string
	// ChannelID is #mod-log (CMDCTRL_DISCORD_MOD_LOG_CHANNEL_ID).
	ChannelID string
}

// Enabled reports whether the channel and the token are set.
func (m ModLog) Enabled() bool { return m.Token != "" && m.ChannelID != "" }

// MissingToken reports the one supported half-configured state: a
// channel but no bot token on this box (the token is provisioned on
// production only, like every other Discord feed's).
func (m ModLog) MissingToken() bool { return m.Token == "" && m.ChannelID != "" }

// ModLogFromEnv reads the mod log's configuration: the channel unset is
// off with no error; one that is not a snowflake is an error naming the
// variable (the boot stops); the channel without a bot token is a ModLog
// whose Enabled() is false and MissingToken() true. It needs no guild:
// the channel knows its own.
func ModLogFromEnv(getenv func(string) string) (ModLog, error) {
	channel := strings.TrimSpace(getenv(ModLogChannelEnv))
	if channel == "" {
		return ModLog{}, nil
	}
	if !isSnowflake(channel) {
		return ModLog{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", ModLogChannelEnv)
	}
	return ModLog{Token: BotFromEnv(getenv).Token, ChannelID: channel}, nil
}

// ModLogField is one line of an entry: a label and its value, the value
// shown as typed.
type ModLogField struct {
	Name, Value string
}

// ModLogEntry is one mod log post. The lobby fills it.
type ModLogEntry struct {
	// Title says what happened: "Player report", "Table ended", ...
	Title string
	// Link is where an admin reads more (the admin view's page); "" for
	// none. Only an absolute http(s) URL is kept.
	Link string
	// Fields are the details, in order. Values are escaped on render.
	Fields []ModLogField
	// Actor names who did it, for the footer: an admin's display name,
	// or a description of the credential. "" for an entry nobody did by
	// hand (a player report says who filed it nowhere: see the ADR).
	Actor string
	// Footer, when set, replaces the "By <actor>" footer.
	Footer string
	// At is when it happened: the embed's timestamp. Zero: none.
	At time.Time
}

// Render builds the message body. Nothing in it can mention anyone:
// allowed_mentions parses nothing, and every value is escaped.
func (e ModLogEntry) Render() lfgMessage {
	embed := lfgEmbed{
		Title:  cutRunes(oneLine(e.Title), maxEmbedTitle),
		URL:    httpURL(e.Link),
		Color:  modLogColor,
		Fields: []lfgEmbedField{},
	}
	if embed.Title == "" {
		embed.Title = "Mod log"
	}
	for _, f := range e.Fields {
		if len(embed.Fields) == maxModLogFields {
			break
		}
		v := escapeText(f.Value, true)
		if v == "" {
			continue
		}
		embed.Fields = append(embed.Fields, lfgEmbedField{
			Name:  cutRunes(oneLine(f.Name), maxEmbedTitle),
			Value: cutRunes(v, maxFieldValue),
			// A reason runs long; everything else is a short line.
			Inline: !strings.Contains(v, "\n") && len([]rune(v)) <= 40,
		})
	}
	switch {
	case e.Footer != "":
		embed.Footer = lfgEmbedFooter{Text: cutRunes(oneLine(e.Footer), maxFieldValue)}
	case e.Actor != "":
		embed.Footer = lfgEmbedFooter{Text: cutRunes("By "+oneLine(e.Actor), maxFieldValue)}
	}
	if !e.At.IsZero() {
		embed.Timestamp = e.At.UTC().Format(time.RFC3339)
	}
	return lfgMessage{Embeds: []lfgEmbed{embed}, AllowedMentions: allowedMentions{Parse: []string{}}}
}

// Post sends one entry to #mod-log.
func (m ModLog) Post(ctx context.Context, client *http.Client, e ModLogEntry) error {
	if !m.Enabled() {
		return ErrModLogNotConfigured
	}
	err := botREST(ctx, client, m.Token, modLogAuditReason, http.MethodPost,
		"/channels/"+m.ChannelID+"/messages", e.Render(), nil, nil)
	var nf *notFoundError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrVoiceForbidden):
		return ErrModLogForbidden
	case errors.As(err, &nf):
		return fmt.Errorf("%w (code %d)", ErrModLogChannelNotFound, nf.code)
	default:
		return err
	}
}
