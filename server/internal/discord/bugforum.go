package discord

// bugforum.go is ADR 0196: the REST calls behind the bug-report mirror,
// a thread in a Discord forum channel for each in-app report that was
// filed as a GitHub issue.
//
// Same rules as lfg.go and results.go, whose helpers it reuses: the
// SERVER calls Discord with the bot token (not the gateway bot binary),
// the token appears in the Authorization header and nowhere else, and
// Discord's own `message` text is never forwarded.
//
// Nothing here decides WHICH reports are mirrored or WHEN a thread
// closes. That is the lobby (lobby/bugforum.go); this file only says
// what a thread looks like on the wire and sends it.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// BugForumEnv names the forum channel reports are mirrored to (ADR 0196
// §1). Unset: nothing is posted.
const BugForumEnv = "CMDCTRL_DISCORD_BUG_FORUM_ID"

// bugForumAuditReason labels the changes in the server's audit log.
const bugForumAuditReason = "CMD & CTRL bug report mirror"

// bugForumColor is the embed's side bar: Discord's red, so a report
// reads differently from an #lfg post's blurple and a result's green.
const bugForumColor = 0xED4245

// Discord's limits a thread stays inside.
const (
	maxThreadName   = 100
	maxBugForumDesc = 3500
	maxMessageText  = 2000
)

// discordForumChannelType is Discord's GUILD_FORUM channel type.
const discordForumChannelType = 15

// maxActiveThreadsBody caps how much of the guild's active-thread
// listing is read. The listing has no paging and every thread in the
// server is in it; botREST's default would cut it off at about 25.
const maxActiveThreadsBody = 2 << 20

// threadAutoArchiveMinutes is how long an idle thread stays open before
// Discord hides it: seven days, Discord's longest standard setting.
const threadAutoArchiveMinutes = 10080

// Errors the forum calls return, beside voice.go's ErrVoiceRateLimited
// (a *RateLimitError), ErrVoiceUnavailable and ErrBotUnauthorized, which
// mean the same thing for a thread. Text is safe to log.
var (
	// ErrBugForumNotConfigured: the mirror is off on this deployment.
	ErrBugForumNotConfigured = errors.New("discord: the bug report mirror is off — " + BugForumEnv + " and " + BotTokenEnv + " must both be set")
	// ErrBugForumForbidden is Discord's 403: the bot cannot post in the
	// forum or manage its threads (ADR 0196's owner setup lists the
	// permissions).
	ErrBugForumForbidden = errors.New("discord: the bot cannot use the bug forum — check View Channel, Send Messages, Create Posts, Send Messages in Posts and Manage Threads there")
	// ErrBugForumNotFound is a 404: the channel id is wrong, the bot is
	// not in the server, or a thread was deleted.
	ErrBugForumNotFound = errors.New("discord: the bug forum or its thread was not found — is the bot in the server?")
	// ErrBugForumNotAForum: the id names a channel that is not a forum.
	ErrBugForumNotAForum = errors.New("discord: " + BugForumEnv + " does not name a forum channel")
)

// BugForum is the mirror's configuration. The zero value is "off", and
// every call on it refuses without touching the network.
type BugForum struct {
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN).
	Token string
	// ChannelID is the forum channel (CMDCTRL_DISCORD_BUG_FORUM_ID).
	ChannelID string
}

// Enabled reports whether the forum and the token are set.
func (f BugForum) Enabled() bool { return f.Token != "" && f.ChannelID != "" }

// MissingToken reports the one supported half-configured state: a forum
// but no bot token on this box (the token is provisioned on production
// only, like every other Discord feed's).
func (f BugForum) MissingToken() bool { return f.Token == "" && f.ChannelID != "" }

// BugForumFromEnv reads ADR 0196 §1's configuration: the forum unset is
// off with no error; one that is not a snowflake is an error naming the
// variable (the boot stops); the forum without a bot token is a BugForum
// whose Enabled() is false and MissingToken() true.
//
// It needs no guild id: the forum channel knows its own guild, and the
// close sweep reads it from there.
func BugForumFromEnv(getenv func(string) string) (BugForum, error) {
	channel := strings.TrimSpace(getenv(BugForumEnv))
	if channel == "" {
		return BugForum{}, nil
	}
	if !isSnowflake(channel) {
		return BugForum{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", BugForumEnv)
	}
	return BugForum{Token: BotFromEnv(getenv).Token, ChannelID: channel}, nil
}

// BugThread is what a mirrored report says. The lobby fills it from the
// text it already filed on GitHub, redacted the same way.
type BugThread struct {
	// IssueNumber and IssueURL are the GitHub issue the report became.
	IssueNumber int
	IssueURL    string
	// Title is the report's title, already redacted. Kind is its noun:
	// "bug", "idea" or "question".
	Title, Kind string
	// Description is the reporter's text, already redacted.
	Description string
	// ReporterID is the reporter's Discord user id, "" for a guest. It
	// is the only person the post may notify.
	ReporterID string
	// Images are the absolute URLs of the screenshots the reporter
	// attached, in order.
	Images []string
}

// ThreadName is the forum post's title: "#123 Title". The number leads
// because the close sweep reads it back (IssueNumberFromThreadName).
func (t BugThread) ThreadName() string {
	title := oneLine(t.Title)
	if title == "" {
		title = "Report"
	}
	return cutRunes(fmt.Sprintf("#%d %s", t.IssueNumber, title), maxThreadName)
}

// oneLine collapses every run of whitespace, line breaks included, to a
// single space. A forum post's name is one line.
func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

var threadNameIssue = regexp.MustCompile(`^#([0-9]{1,9})(?: |$)`)

// IssueNumberFromThreadName reads the GitHub issue number out of a name
// ThreadName made, 0 for any other name.
func IssueNumberFromThreadName(name string) int {
	m := threadNameIssue.FindStringSubmatch(name)
	if m == nil {
		return 0
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0
	}
	return n
}

// bugThreadBody is a forum post's create body: the thread and its
// first message.
type bugThreadBody struct {
	Name                string     `json:"name"`
	AutoArchiveDuration int        `json:"auto_archive_duration"`
	Message             lfgMessage `json:"message"`
}

// Render builds the body for a new forum post.
//
// The reporter is the only mention: a user mention in the content with
// allowed_mentions naming exactly that one id, so nothing in the
// reporter's own text (every character Discord reads as markup or a
// mention is escaped, as results.go's names are) can ping anyone else,
// whatever it says. A reporter id that is not a snowflake is dropped.
func (t BugThread) Render() bugThreadBody {
	desc := escapeText(t.Description, true)
	if desc == "" {
		desc = "_no description provided_"
	}
	embed := lfgEmbed{
		Title:       cutRunes(fmt.Sprintf("GitHub issue #%d", t.IssueNumber), maxEmbedTitle),
		URL:         httpURL(t.IssueURL),
		Color:       bugForumColor,
		Description: cutRunes(desc, maxBugForumDesc),
		Fields:      []lfgEmbedField{},
		Footer:      lfgEmbedFooter{Text: "Reported from the app"},
	}
	if t.Kind != "" {
		embed.Fields = append(embed.Fields, lfgEmbedField{Name: "Kind", Value: cutRunes(escapeText(t.Kind, false), maxFieldValue), Inline: true})
	}
	if u := httpURL(t.IssueURL); u != "" {
		embed.Fields = append(embed.Fields, lfgEmbedField{Name: "Tracking issue", Value: u, Inline: true})
	}
	var shots []string
	for _, u := range t.Images {
		if u = httpURL(u); u != "" {
			shots = append(shots, u)
		}
	}
	if len(shots) > 0 {
		embed.Image = &lfgEmbedImage{URL: shots[0]}
		embed.Fields = append(embed.Fields, lfgEmbedField{Name: "Screenshots", Value: cutRunes(strings.Join(shots, "\n"), maxFieldValue)})
	}

	msg := lfgMessage{Embeds: []lfgEmbed{embed}, AllowedMentions: allowedMentions{Parse: []string{}}}
	content := "A maintainer may reply in this thread."
	if isSnowflake(t.ReporterID) {
		content = "<@" + t.ReporterID + "> thanks for the report. A maintainer may reply in this thread."
		msg.AllowedMentions.Users = []string{t.ReporterID}
	}
	msg.Content = &content
	return bugThreadBody{Name: t.ThreadName(), AutoArchiveDuration: threadAutoArchiveMinutes, Message: msg}
}

// httpURL returns u when it is an absolute http(s) URL and "" otherwise,
// so an embed never carries a scheme Discord would reject.
func httpURL(u string) string {
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return u
	}
	return ""
}

// forumError gives the voice-worded errors a forum's meaning.
func (f BugForum) forumError(err error) error {
	var nf *notFoundError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrVoiceForbidden):
		return ErrBugForumForbidden
	case errors.As(err, &nf):
		return fmt.Errorf("%w (code %d)", ErrBugForumNotFound, nf.code)
	default:
		return err
	}
}

// CreateThread opens a forum post for a report and returns the thread's
// id.
func (f BugForum) CreateThread(ctx context.Context, client *http.Client, t BugThread) (string, error) {
	if !f.Enabled() {
		return "", ErrBugForumNotConfigured
	}
	var out struct {
		ID string `json:"id"`
	}
	err := botREST(ctx, client, f.Token, bugForumAuditReason, http.MethodPost,
		"/channels/"+f.ChannelID+"/threads", t.Render(), &out, nil)
	if err != nil {
		return "", f.forumError(err)
	}
	if out.ID == "" {
		return "", errors.New("discord: thread response carried no id")
	}
	return out.ID, nil
}

// ForumThread is one open post the bot made in the forum.
type ForumThread struct {
	ID   string
	Name string
}

// ActiveThreads lists the open posts in the forum that the bot itself
// created. A post a person made, or one a person renamed away from the
// "#123 Title" form, is not in the list: the mirror only ever closes its
// own.
func (f BugForum) ActiveThreads(ctx context.Context, client *http.Client) ([]ForumThread, error) {
	if !f.Enabled() {
		return nil, ErrBugForumNotConfigured
	}
	var ch struct {
		GuildID string `json:"guild_id"`
		Type    int    `json:"type"`
	}
	if err := botREST(ctx, client, f.Token, bugForumAuditReason, http.MethodGet, "/channels/"+f.ChannelID, nil, &ch, nil); err != nil {
		return nil, f.forumError(err)
	}
	if ch.Type != discordForumChannelType {
		return nil, ErrBugForumNotAForum
	}
	if !isSnowflake(ch.GuildID) {
		return nil, errors.New("discord: forum channel response carried no guild id")
	}
	var me struct {
		ID string `json:"id"`
	}
	if err := botREST(ctx, client, f.Token, bugForumAuditReason, http.MethodGet, "/users/@me", nil, &me, nil); err != nil {
		return nil, f.forumError(err)
	}
	if me.ID == "" {
		return nil, errors.New("discord: bot user response carried no id")
	}
	var list struct {
		Threads []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			ParentID string `json:"parent_id"`
			OwnerID  string `json:"owner_id"`
		} `json:"threads"`
	}
	err := botRESTLimit(ctx, client, f.Token, bugForumAuditReason, http.MethodGet,
		"/guilds/"+ch.GuildID+"/threads/active", nil, &list, nil, maxActiveThreadsBody)
	if err != nil {
		return nil, f.forumError(err)
	}
	var out []ForumThread
	for _, t := range list.Threads {
		if t.ParentID == f.ChannelID && t.OwnerID == me.ID && IssueNumberFromThreadName(t.Name) > 0 {
			out = append(out, ForumThread{ID: t.ID, Name: t.Name})
		}
	}
	return out, nil
}

// PostNote adds a message to a thread. It mentions nobody.
func (f BugForum) PostNote(ctx context.Context, client *http.Client, threadID, text string) error {
	if !f.Enabled() {
		return ErrBugForumNotConfigured
	}
	if !isSnowflake(threadID) {
		return errors.New("discord: not a thread id")
	}
	content := cutRunes(text, maxMessageText)
	body := lfgMessage{Content: &content, Embeds: []lfgEmbed{}, AllowedMentions: allowedMentions{Parse: []string{}}}
	return f.forumError(botREST(ctx, client, f.Token, bugForumAuditReason, http.MethodPost,
		"/channels/"+threadID+"/messages", body, nil, nil))
}

// Archive closes a thread. Anyone who posts in it afterwards reopens it,
// which is the point: the reporter can still answer a late question.
func (f BugForum) Archive(ctx context.Context, client *http.Client, threadID string) error {
	if !f.Enabled() {
		return ErrBugForumNotConfigured
	}
	if !isSnowflake(threadID) {
		return errors.New("discord: not a thread id")
	}
	return f.forumError(botREST(ctx, client, f.Token, bugForumAuditReason, http.MethodPatch,
		"/channels/"+threadID, map[string]any{"archived": true}, nil, nil))
}
