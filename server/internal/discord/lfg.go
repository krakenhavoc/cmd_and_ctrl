package discord

// lfg.go is ADR 0161 §2, §3 and §5: the REST calls behind a table's
// post on the community server's #lfg board. Posting the message,
// editing it as the table changes, and deleting it.
//
// Same rules as voice.go and dm.go: the SERVER calls Discord with the
// bot token, the token appears in the Authorization header and nowhere
// else, and Discord's own `message` text is never forwarded.
//
// Nothing here decides WHEN a table is posted. That is the lobby's LFG
// manager (lfg_manager.go); this file only says what a post looks like
// on the wire and sends it.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// The two variables ADR 0161 §7 adds.
const (
	// LFGChannelEnv names the #lfg channel tables are posted to.
	LFGChannelEnv = "CMDCTRL_DISCORD_LFG_CHANNEL_ID"
	// LFGRoleEnv names the role a new post pings. Optional.
	LFGRoleEnv = "CMDCTRL_DISCORD_LFG_ROLE_ID"
)

// lfgAuditReason labels the deletes in the server's audit log.
const lfgAuditReason = "CMD & CTRL #lfg post"

// Discord's limits a post stays inside.
const (
	maxEmbedTitle = 256
	maxFieldValue = 1024
)

// The message shapes a post uses (Discord's component and embed types).
const (
	componentActionRow = 1
	componentButton    = 2
	buttonStyleLink    = 5
	// lfgColor is the embed's side bar: Discord's blurple, which the
	// client's voice chip uses too.
	lfgColor = 0x5865F2
)

// Discord's error codes this file reads, beside voice.go's.
const codeUnknownMessage = 10008

// Errors the LFG calls return, beside voice.go's ErrVoiceRateLimited
// (a *RateLimitError), ErrVoiceUnavailable and ErrBotUnauthorized,
// which mean the same thing for a message. Text is safe to log.
var (
	// ErrLFGNotConfigured: the board is off on this deployment.
	ErrLFGNotConfigured = errors.New("discord: the #lfg board is off — " + LFGChannelEnv + ", " +
		CommunityGuildEnv + " and " + BotTokenEnv + " must all be set")
	// ErrLFGForbidden is Discord's 403: the bot cannot post in the
	// channel (ADR 0161's owner setup lists the permissions).
	ErrLFGForbidden = errors.New("discord: the bot cannot post in the #lfg channel — check View Channel, Send Messages and Embed Links there")
	// ErrLFGMessageGone is a 404 for the message: someone deleted the
	// post on Discord.
	ErrLFGMessageGone = errors.New("discord: the #lfg post no longer exists")
	// ErrLFGChannelNotFound is a 404 for anything else: the channel id
	// is wrong, or the bot is not in the server.
	ErrLFGChannelNotFound = errors.New("discord: the #lfg channel was not found — is the bot in the server?")
)

// LFG is the #lfg board's configuration. The zero value is "off", and
// every call on it refuses without touching the network.
type LFG struct {
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN).
	Token string
	// GuildID is the community server (CMDCTRL_DISCORD_COMMUNITY_GUILD_ID).
	GuildID string
	// ChannelID is #lfg (CMDCTRL_DISCORD_LFG_CHANNEL_ID).
	ChannelID string
	// RoleID is the role a new post pings (CMDCTRL_DISCORD_LFG_ROLE_ID);
	// "" never pings.
	RoleID string
}

// Enabled reports whether the channel, the guild and the token are set.
func (f LFG) Enabled() bool { return f.Token != "" && f.GuildID != "" && f.ChannelID != "" }

// MissingToken reports the one supported half-configured state: a
// channel and a guild, but no bot token on this box.
func (f LFG) MissingToken() bool { return f.Token == "" && f.ChannelID != "" }

// LFGFromEnv reads ADR 0161 §7's configuration.
//
//   - Channel and role unset: off, no error.
//   - A role without the channel, the channel without the community
//     guild, or any of the three not a snowflake: an error naming the
//     variable, and the boot stops (GateFromEnv, VoiceFromEnv).
//   - Channel and guild set, no bot token: an LFG whose Enabled() is
//     false and whose MissingToken() is true. The token is production
//     only, so a box that inherits the variables must still boot; main
//     says once that the board is off and why.
func LFGFromEnv(getenv func(string) string) (LFG, error) {
	channel := strings.TrimSpace(getenv(LFGChannelEnv))
	role := strings.TrimSpace(getenv(LFGRoleEnv))
	if channel == "" {
		if role != "" {
			return LFG{}, fmt.Errorf("%s is set but %s is not: the role is pinged by posts in that channel, so name it", LFGRoleEnv, LFGChannelEnv)
		}
		return LFG{}, nil
	}
	if !isSnowflake(channel) {
		return LFG{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", LFGChannelEnv)
	}
	if role != "" && !isSnowflake(role) {
		return LFG{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", LFGRoleEnv)
	}
	guild := strings.TrimSpace(getenv(CommunityGuildEnv))
	if guild == "" {
		return LFG{}, fmt.Errorf("%s is set but %s is not: #lfg is a channel in the community server, so name it", LFGChannelEnv, CommunityGuildEnv)
	}
	if !isSnowflake(guild) {
		return LFG{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", CommunityGuildEnv)
	}
	return LFG{
		Token:     BotFromEnv(getenv).Token,
		GuildID:   guild,
		ChannelID: channel,
		RoleID:    role,
	}, nil
}

// LFGField is one line of a post's embed: a table setting that is not
// the default (ADR 0161 §2).
type LFGField struct {
	Name, Value string
}

// LFGPost is what a table's post says. The lobby fills it; Render turns
// it into Discord's message body.
type LFGPost struct {
	// Table is the table's name, already through the name filter.
	Table string
	// HostID is the host's Discord user id, mentioned in the embed.
	HostID string
	// Seated and Seats are the seat count: "2 of 4".
	Seated, Seats int
	// Voice is whether the table has asked for a voice room (ADR 0159).
	Voice bool
	// Settings are the table settings that change how the game plays,
	// in display order, only when not the default.
	Settings []LFGField
	// CreatedAt is when the table was opened: the embed's timestamp.
	CreatedAt time.Time
	// JoinURL is the table's invite link. "" when this process does
	// not hold it (lost to a restart): an edit then leaves the button
	// as it is, and Post refuses.
	JoinURL string
}

// Full reports whether every seat is taken.
func (p LFGPost) Full() bool { return p.Seats > 0 && p.Seated >= p.Seats }

// lfgEmbed and the types below are Discord's message JSON, the part a
// post uses.
type lfgEmbed struct {
	Title string `json:"title"`
	// Description is the text under the title. Only the results post
	// (results.go) uses it.
	Description string          `json:"description,omitempty"`
	Color       int             `json:"color"`
	Fields      []lfgEmbedField `json:"fields"`
	Footer      lfgEmbedFooter  `json:"footer"`
	Timestamp   string          `json:"timestamp,omitempty"`
}

type lfgEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type lfgEmbedFooter struct {
	Text string `json:"text"`
}

type lfgComponent struct {
	Type       int            `json:"type"`
	Components []lfgComponent `json:"components,omitempty"`
	Style      int            `json:"style,omitempty"`
	Label      string         `json:"label,omitempty"`
	URL        string         `json:"url,omitempty"`
	Disabled   bool           `json:"disabled,omitempty"`
}

type allowedMentions struct {
	// Parse is always present and empty: nothing in the content is
	// parsed for a mention except what Roles names.
	Parse []string `json:"parse"`
	Roles []string `json:"roles,omitempty"`
}

// lfgMessage is a create or edit body. Content and Components are
// pointers so an edit can leave them as they are by omitting them.
type lfgMessage struct {
	Content         *string         `json:"content,omitempty"`
	Embeds          []lfgEmbed      `json:"embeds"`
	Components      *[]lfgComponent `json:"components,omitempty"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

// render builds the body for a post. ping adds the role mention to the
// content and lets exactly that role be notified; edit leaves the
// content alone (an edit never pings, ADR 0161 §3) and, without a join
// URL, the button too.
func (f LFG) render(p LFGPost, ping, edit bool) lfgMessage {
	seats := fmt.Sprintf("%d of %d", p.Seated, p.Seats)
	if p.Full() {
		seats += " · Full"
	}
	voice := "No"
	if p.Voice {
		voice = "Yes"
	}
	fields := []lfgEmbedField{
		{Name: "Host", Value: "<@" + p.HostID + ">", Inline: true},
		{Name: "Seats", Value: seats, Inline: true},
		{Name: "Voice room", Value: voice, Inline: true},
	}
	for _, s := range p.Settings {
		fields = append(fields, lfgEmbedField{Name: s.Name, Value: cutRunes(s.Value, maxFieldValue), Inline: true})
	}
	embed := lfgEmbed{
		Title:  PostTitle(p.Table),
		Color:  lfgColor,
		Fields: fields,
		Footer: lfgEmbedFooter{Text: "Table opened"},
	}
	if !p.CreatedAt.IsZero() {
		embed.Timestamp = p.CreatedAt.UTC().Format(time.RFC3339)
	}
	msg := lfgMessage{Embeds: []lfgEmbed{embed}, AllowedMentions: allowedMentions{Parse: []string{}}}
	if !edit {
		content := ""
		if ping && f.RoleID != "" {
			content = "<@&" + f.RoleID + ">"
			msg.AllowedMentions.Roles = []string{f.RoleID}
		}
		msg.Content = &content
	}
	if p.JoinURL != "" {
		label := "Join table"
		if p.Full() {
			label = "Table full"
		}
		row := []lfgComponent{{
			Type: componentActionRow,
			Components: []lfgComponent{{
				Type: componentButton, Style: buttonStyleLink,
				Label: label, URL: p.JoinURL, Disabled: p.Full(),
			}},
		}}
		msg.Components = &row
	}
	return msg
}

// PostTitle is the embed title for a table: its name, trimmed, cut to
// Discord's 256 characters. A blank name falls back to "Table".
func PostTitle(table string) string {
	name := strings.TrimSpace(table)
	if name == "" {
		return "Table"
	}
	return cutRunes(name, maxEmbedTitle)
}

func cutRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n]))
}

// Post creates a table's message in #lfg and returns its id. ping
// mentions the configured role (no role configured: no ping). A post
// needs the join URL: a message with no way to join is not posted.
func (f LFG) Post(ctx context.Context, client *http.Client, p LFGPost, ping bool) (string, error) {
	if p.JoinURL == "" {
		return "", errors.New("discord: an #lfg post needs the table's join link")
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := f.call(ctx, client, http.MethodPost, "/channels/"+f.ChannelID+"/messages", f.render(p, ping, false), &out, nil); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", errors.New("discord: message response carried no id")
	}
	return out.ID, nil
}

// Edit re-renders a table's message. It never pings. ErrLFGMessageGone
// when the message was deleted on Discord.
func (f LFG) Edit(ctx context.Context, client *http.Client, messageID string, p LFGPost) error {
	return f.call(ctx, client, http.MethodPatch, "/channels/"+f.ChannelID+"/messages/"+messageID, f.render(p, false, true), nil, nil)
}

// Delete removes a table's message. One that is already gone is
// success.
func (f LFG) Delete(ctx context.Context, client *http.Client, messageID string) error {
	return f.call(ctx, client, http.MethodDelete, "/channels/"+f.ChannelID+"/messages/"+messageID, nil, nil,
		[]int{codeUnknownMessage})
}

// call sends one request through botREST and gives the voice-worded
// errors a message's meaning.
func (f LFG) call(ctx context.Context, client *http.Client, method, path string, body, out any, goneOK []int) error {
	if !f.Enabled() {
		return ErrLFGNotConfigured
	}
	err := botREST(ctx, client, f.Token, lfgAuditReason, method, path, body, out, goneOK)
	var nf *notFoundError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrVoiceForbidden):
		return ErrLFGForbidden
	case errors.As(err, &nf) && nf.code == codeUnknownMessage:
		return ErrLFGMessageGone
	case errors.As(err, &nf):
		return fmt.Errorf("%w (code %d)", ErrLFGChannelNotFound, nf.code)
	default:
		return err
	}
}
