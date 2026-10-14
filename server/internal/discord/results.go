package discord

// results.go is ADR 0174 §2, §5 and §6: the REST call behind a finished
// game's post in the community server's #game-results channel.
//
// Same rules as lfg.go, whose helpers it reuses: the SERVER calls
// Discord with the bot token, the token appears in the Authorization
// header and nowhere else, and Discord's own `message` text is never
// forwarded. Unlike an #lfg post a results post is made once and never
// edited or deleted, and it mentions nobody: every name in it is plain
// text and allowed_mentions parses nothing.
//
// Nothing here decides WHETHER a game is posted. That is the lobby
// (lobby/results_feed.go); this file only says what a post looks like
// on the wire and sends it.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ResultsChannelEnv names the #game-results channel (ADR 0174 §6).
const ResultsChannelEnv = "CMDCTRL_DISCORD_RESULTS_CHANNEL_ID"

// resultsAuditReason labels the posts in the server's audit log.
const resultsAuditReason = "CMD & CTRL game results"

// resultsColor is the embed's side bar: Discord's green, so a result
// reads differently from an #lfg post's blurple.
const resultsColor = 0x57F287

// Errors the results call returns, beside voice.go's ErrVoiceRateLimited
// (a *RateLimitError), ErrVoiceUnavailable and ErrBotUnauthorized, which
// mean the same thing for a message. Text is safe to log.
var (
	// ErrResultsNotConfigured: the feed is off on this deployment.
	ErrResultsNotConfigured = errors.New("discord: the results feed is off — " + ResultsChannelEnv + ", " +
		CommunityGuildEnv + " and " + BotTokenEnv + " must all be set")
	// ErrResultsForbidden is Discord's 403: the bot cannot post in the
	// channel (ADR 0174's owner setup lists the permissions).
	ErrResultsForbidden = errors.New("discord: the bot cannot post in the results channel — check View Channel, Send Messages and Embed Links there")
	// ErrResultsChannelNotFound is a 404: the channel id is wrong, or
	// the bot is not in the server.
	ErrResultsChannelNotFound = errors.New("discord: the results channel was not found — is the bot in the server?")
)

// Results is the results feed's configuration. The zero value is "off",
// and every call on it refuses without touching the network.
type Results struct {
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN).
	Token string
	// GuildID is the community server (CMDCTRL_DISCORD_COMMUNITY_GUILD_ID).
	GuildID string
	// ChannelID is #game-results (CMDCTRL_DISCORD_RESULTS_CHANNEL_ID).
	ChannelID string
}

// Enabled reports whether the channel, the guild and the token are set.
func (r Results) Enabled() bool { return r.Token != "" && r.GuildID != "" && r.ChannelID != "" }

// MissingToken reports the one supported half-configured state: a
// channel and a guild, but no bot token on this box.
func (r Results) MissingToken() bool { return r.Token == "" && r.ChannelID != "" }

// ResultsFromEnv reads ADR 0174 §6's configuration, with LFGFromEnv's
// rules: the channel unset is off with no error; the channel without the
// community guild, or either not a snowflake, is an error naming the
// variable; the channel and guild without a bot token is a Results whose
// Enabled() is false and MissingToken() true.
func ResultsFromEnv(getenv func(string) string) (Results, error) {
	channel := strings.TrimSpace(getenv(ResultsChannelEnv))
	if channel == "" {
		return Results{}, nil
	}
	if !isSnowflake(channel) {
		return Results{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", ResultsChannelEnv)
	}
	guild := strings.TrimSpace(getenv(CommunityGuildEnv))
	if guild == "" {
		return Results{}, fmt.Errorf("%s is set but %s is not: the results channel is in the community server, so name it", ResultsChannelEnv, CommunityGuildEnv)
	}
	if !isSnowflake(guild) {
		return Results{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", CommunityGuildEnv)
	}
	return Results{Token: BotFromEnv(getenv).Token, GuildID: guild, ChannelID: channel}, nil
}

// ResultsSeat is one seat's line of a post.
type ResultsSeat struct {
	// Name is the seat's name as the table shows it.
	Name string
	// Kind marks a seat that is not a person: "AI agent" or "bot". ""
	// for a person.
	Kind string
	// Commanders are the seat's commander names.
	Commanders []string
	// Place is 1 for the winner, then by reverse elimination order; 0
	// when the game gave it none.
	Place int
	// Result is "win", "loss", "draw" or "".
	Result string
	// Conceded and EliminatedTurn say how the seat left; the turn is 0
	// for a seat that was never eliminated.
	Conceded       bool
	EliminatedTurn int
}

// ResultsPost is what a finished game's post says. The lobby fills it;
// Render turns it into Discord's message body.
type ResultsPost struct {
	// Table is the table's name.
	Table string
	// Seats are in placing order, the winner first.
	Seats []ResultsSeat
	// Turns is the game's last turn number.
	Turns int
	// Duration is from the game starting (or the table opening, when the
	// start is not on record) to its end; zero is left out.
	Duration time.Duration
	// EndedAt is the embed's timestamp.
	EndedAt time.Time
}

// Render builds the message body for a post. It carries no content line
// and no mention, and allowed_mentions parses nothing, so a seat named
// "@everyone" is only text.
func (p ResultsPost) Render() lfgMessage {
	embed := lfgEmbed{
		Title:       PostTitle(p.Table),
		Color:       resultsColor,
		Description: p.verdict(),
		Fields:      []lfgEmbedField{},
		Footer:      lfgEmbedFooter{Text: "Game ended"},
	}
	for _, s := range p.Seats {
		name := plainText(s.Name)
		if name == "" {
			name = "Seat"
		}
		if s.Kind != "" {
			name += " (" + s.Kind + ")"
		}
		if s.Place > 0 {
			name = ordinal(s.Place) + " · " + name
		}
		value := "No commander"
		if len(s.Commanders) > 0 {
			names := make([]string, len(s.Commanders))
			for i, c := range s.Commanders {
				names[i] = plainText(c)
			}
			value = strings.Join(names, " / ")
		}
		switch {
		case s.Conceded && s.EliminatedTurn > 0:
			value += fmt.Sprintf("\nConceded on turn %d", s.EliminatedTurn)
		case s.Conceded:
			value += "\nConceded"
		case s.EliminatedTurn > 0:
			value += fmt.Sprintf("\nEliminated on turn %d", s.EliminatedTurn)
		}
		embed.Fields = append(embed.Fields, lfgEmbedField{
			Name: cutRunes(name, maxEmbedTitle), Value: cutRunes(value, maxFieldValue),
		})
	}
	embed.Fields = append(embed.Fields, lfgEmbedField{Name: "Turns", Value: fmt.Sprint(p.Turns), Inline: true})
	if d := formatDuration(p.Duration); d != "" {
		embed.Fields = append(embed.Fields, lfgEmbedField{Name: "Duration", Value: d, Inline: true})
	}
	if !p.EndedAt.IsZero() {
		embed.Timestamp = p.EndedAt.UTC().Format(time.RFC3339)
	}
	return lfgMessage{Embeds: []lfgEmbed{embed}, AllowedMentions: allowedMentions{Parse: []string{}}}
}

// verdict is the embed's description: the winner, or the draw.
func (p ResultsPost) verdict() string {
	var winners, drawn []string
	for _, s := range p.Seats {
		label := plainText(s.Name)
		if s.Kind != "" {
			label += " (" + s.Kind + ")"
		}
		switch s.Result {
		case "win":
			winners = append(winners, label)
		case "draw":
			drawn = append(drawn, label)
		}
	}
	switch {
	case len(winners) > 0:
		return "**Winner:** " + strings.Join(winners, ", ")
	case len(drawn) > 0:
		return "**Draw** between " + strings.Join(drawn, ", ")
	default:
		return "**Draw**"
	}
}

func ordinal(n int) string {
	switch {
	case n%100 >= 11 && n%100 <= 13:
		return fmt.Sprintf("%dth", n)
	case n%10 == 1:
		return fmt.Sprintf("%dst", n)
	case n%10 == 2:
		return fmt.Sprintf("%dnd", n)
	case n%10 == 3:
		return fmt.Sprintf("%drd", n)
	}
	return fmt.Sprintf("%dth", n)
}

// formatDuration is "1h 12m", "42m" or "under a minute"; "" for none.
func formatDuration(d time.Duration) string {
	switch {
	case d <= 0:
		return ""
	case d < time.Minute:
		return "under a minute"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
}

// plainText makes a name safe to show as text: surrounding space is
// trimmed and every character Discord reads as markup or a mention is
// escaped, so a seat called "<@123>" or "**x**" shows as typed.
func plainText(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\', '*', '_', '~', '`', '|', '>', '<', '@', '#', '[', ']', ':':
			b.WriteByte('\\')
		case '\n', '\r':
			r = ' '
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Post sends a finished game's message to #game-results.
func (r Results) Post(ctx context.Context, client *http.Client, p ResultsPost) error {
	if !r.Enabled() {
		return ErrResultsNotConfigured
	}
	err := botREST(ctx, client, r.Token, resultsAuditReason, http.MethodPost,
		"/channels/"+r.ChannelID+"/messages", p.Render(), nil, nil)
	var nf *notFoundError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrVoiceForbidden):
		return ErrResultsForbidden
	case errors.As(err, &nf):
		return fmt.Errorf("%w (code %d)", ErrResultsChannelNotFound, nf.code)
	default:
		return err
	}
}
