package discord

// voice.go is ADR 0159 §4: the REST calls behind a table's private
// voice room in the community server. Creating the room, letting one
// player in or out, deleting it, listing what the category holds, and
// the two questions the voice manager asks about a player (are they in
// the server, which voice channel are they connected to).
//
// Like dm.go this is the SERVER calling Discord with the bot token, not
// the gateway bot, and the same rule holds: the token appears in the
// Authorization header and nowhere else. Discord's own `message` text
// is never forwarded; errors are this file's sentinels.
//
// Nothing here decides WHEN a room exists or who belongs in it. That is
// the lobby's voice manager (ADR 0159 slice 3); these are the calls it
// makes.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// VoiceCategoryEnv names the category in the community server that
// table rooms are created in (ADR 0159 §7).
const VoiceCategoryEnv = "CMDCTRL_DISCORD_VOICE_CATEGORY_ID"

// VoicePullInEnv turns on moving seated players who are in another voice
// channel into their table's room (#3311). Off unless it is a true value.
const VoicePullInEnv = "CMDCTRL_DISCORD_VOICE_PULL_IN"

// Discord permission bits a room's overwrites use. Discord sends and
// takes permission sets as decimal strings of these bits.
const (
	PermManageChannels     int64 = 1 << 4
	PermStream             int64 = 1 << 9 // "Video" in the client
	PermViewChannel        int64 = 1 << 10
	PermSendMessages       int64 = 1 << 11
	PermReadMessageHistory int64 = 1 << 16
	PermConnect            int64 = 1 << 20
	PermSpeak              int64 = 1 << 21
	PermMoveMembers        int64 = 1 << 24
	PermUseVAD             int64 = 1 << 25
)

// The three overwrites ADR 0159 §3 gives a room.
const (
	// EveryoneDeny hides the room from the server: the @everyone role
	// (whose id is the guild's) may neither see nor join it.
	EveryoneDeny = PermViewChannel | PermConnect
	// PlayerAllow is what a seated player who is a server member gets:
	// in, heard, seen, and the room's text chat.
	PlayerAllow = PermViewChannel | PermConnect | PermSpeak | PermStream |
		PermUseVAD | PermSendMessages | PermReadMessageHistory
	// BotAllow keeps the bot able to see the room it made despite the
	// @everyone deny. Managing channels and permissions and moving
	// members come from the bot's role, server-wide: Discord refuses an
	// overwrite that grants Manage Permissions to anyone but an
	// administrator, so the bot cannot hand them to itself here.
	BotAllow = PermViewChannel | PermConnect
)

// Overwrite target types.
const (
	overwriteRole   = 0
	overwriteMember = 1
)

// channelTypeVoice is Discord's GUILD_VOICE.
const channelTypeVoice = 2

// maxRoomName is Discord's channel-name limit, in characters.
const maxRoomName = 100

// auditReason labels every change in the server's audit log.
const auditReason = "CMD & CTRL table voice room"

// Discord's JSON error codes this file tells apart. A 404 alone does
// not say WHAT was unknown; the code does.
const (
	codeUnknownChannel    = 10003
	codeUnknownGuild      = 10004
	codeUnknownMember     = 10007
	codeUnknownOverwrite  = 10009
	codeUnknownUser       = 10013
	codeUnknownVoiceState = 10065
)

// Errors the voice calls return. Text is safe to log and to show an
// admin; none carries the token or Discord's own wording.
var (
	// ErrVoiceNotConfigured: the feature is off on this deployment.
	ErrVoiceNotConfigured = errors.New("discord: voice rooms are off — " + VoiceCategoryEnv + ", " +
		CommunityGuildEnv + " and " + BotTokenEnv + " must all be set")
	// ErrVoiceForbidden is Discord's 403: the bot lacks a permission
	// the call needs (ADR 0159's owner setup lists them).
	ErrVoiceForbidden = errors.New("discord: the bot lacks a permission for this — check its role in the community server")
	// ErrVoiceNotFound is a 404 the call did not expect: the guild or
	// the category is wrong, or the bot is not in the server.
	ErrVoiceNotFound = errors.New("discord: the community server or the voice category was not found — is the bot in the server?")
	// ErrVoiceRateLimited is Discord's 429. Unwrap a *RateLimitError
	// for how long to wait.
	ErrVoiceRateLimited = errors.New("discord: rate-limited by Discord")
	// ErrVoiceUnavailable is a 5xx or a transport failure: Discord did
	// not answer. Worth retrying.
	ErrVoiceUnavailable = errors.New("discord: Discord did not answer")
)

// RateLimitError is a 429 with the wait Discord asked for.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%v (retry after %s)", ErrVoiceRateLimited, e.RetryAfter)
}

// Is makes errors.Is(err, ErrVoiceRateLimited) hold.
func (e *RateLimitError) Is(target error) bool { return target == ErrVoiceRateLimited }

// Transient reports whether a voice call's error is worth retrying:
// a rate limit or Discord not answering. Every other error (a missing
// permission, a wrong id, a rejected token) will fail the same way
// again.
func Transient(err error) bool {
	return errors.Is(err, ErrVoiceRateLimited) || errors.Is(err, ErrVoiceUnavailable)
}

// Voice is the configuration for table rooms. The zero value is "off",
// and every call on it refuses without touching the network.
type Voice struct {
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN).
	Token string
	// GuildID is the community server (CMDCTRL_DISCORD_COMMUNITY_GUILD_ID).
	GuildID string
	// CategoryID is where rooms go (CMDCTRL_DISCORD_VOICE_CATEGORY_ID).
	CategoryID string
	// PullIn is whether a seated, granted player who is in another voice
	// channel of the server is moved into the table's room
	// (CMDCTRL_DISCORD_VOICE_PULL_IN, default off).
	PullIn bool
}

// Enabled reports whether all three are set.
func (v Voice) Enabled() bool { return v.Token != "" && v.GuildID != "" && v.CategoryID != "" }

// VoiceFromEnv reads ADR 0159 §7's configuration.
//
//   - Category unset: off, no error.
//   - Category malformed, or set without the community guild (or with a
//     malformed guild): an error naming the variable. Rooms in a server
//     nobody named are a misconfiguration, so the boot stops.
//   - Category and guild set, no bot token: a Voice whose Enabled() is
//     false and whose MissingToken() is true. The token is provisioned
//     on production only, so a box that inherits a repo-level category
//     must still boot; main says once that voice is off and why.
func VoiceFromEnv(getenv func(string) string) (Voice, error) {
	category := strings.TrimSpace(getenv(VoiceCategoryEnv))
	if category == "" {
		return Voice{}, nil
	}
	if !isSnowflake(category) {
		return Voice{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", VoiceCategoryEnv)
	}
	guild := strings.TrimSpace(getenv(CommunityGuildEnv))
	if guild == "" {
		return Voice{}, fmt.Errorf("%s is set but %s is not: voice rooms are created in the community server, so name it", VoiceCategoryEnv, CommunityGuildEnv)
	}
	if !isSnowflake(guild) {
		return Voice{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", CommunityGuildEnv)
	}
	return Voice{
		Token:      BotFromEnv(getenv).Token,
		GuildID:    guild,
		CategoryID: category,
		PullIn:     truthy(getenv(VoicePullInEnv)),
	}, nil
}

func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// MissingToken reports the one supported half-configured state: a
// category and a guild, but no bot token on this box.
func (v Voice) MissingToken() bool { return v.Token == "" && v.CategoryID != "" }

// RoomName is the channel name for a table: its name, trimmed, cut to
// Discord's 100 characters. A blank name falls back to "Table".
func RoomName(table string) string {
	name := strings.TrimSpace(table)
	if name == "" {
		return "Table"
	}
	if utf8.RuneCountInString(name) <= maxRoomName {
		return name
	}
	runes := []rune(name)
	return strings.TrimSpace(string(runes[:maxRoomName]))
}

// RoomURL is the link that opens a room in the Discord app or web
// client.
func (v Voice) RoomURL(channelID string) string {
	return "https://discord.com/channels/" + v.GuildID + "/" + channelID
}

// Room is one voice channel in the category.
type Room struct {
	ID   string
	Name string
}

// overwrite is Discord's permission-overwrite object. allow and deny
// are decimal strings.
type overwrite struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Allow string `json:"allow"`
	Deny  string `json:"deny"`
}

func perms(p int64) string { return strconv.FormatInt(p, 10) }

// BotUserID is the bot's own user id, which a room's bot overwrite
// names. GET /users/@me. The caller caches it; it never changes for a
// token.
func (v Voice) BotUserID(ctx context.Context, client *http.Client) (string, error) {
	var me struct {
		ID string `json:"id"`
	}
	if err := v.call(ctx, client, http.MethodGet, "/users/@me", nil, &me, nil); err != nil {
		return "", err
	}
	if me.ID == "" {
		return "", errors.New("discord: /users/@me carried no id")
	}
	return me.ID, nil
}

// CreateRoom makes a private voice channel in the category, named for
// the table, that only botUserID and members can see and join. members
// must already be known to be in the server (IsMember); an overwrite
// does not invite anyone. Returns the new channel's id.
func (v Voice) CreateRoom(ctx context.Context, client *http.Client, name, botUserID string, members []string) (string, error) {
	if botUserID == "" {
		return "", errors.New("discord: CreateRoom needs the bot's user id")
	}
	ows := []overwrite{
		{ID: v.GuildID, Type: overwriteRole, Allow: perms(0), Deny: perms(EveryoneDeny)},
		{ID: botUserID, Type: overwriteMember, Allow: perms(BotAllow), Deny: perms(0)},
	}
	for _, m := range members {
		if m == "" || m == botUserID {
			continue
		}
		ows = append(ows, overwrite{ID: m, Type: overwriteMember, Allow: perms(PlayerAllow), Deny: perms(0)})
	}
	body := map[string]any{
		"name":                  RoomName(name),
		"type":                  channelTypeVoice,
		"parent_id":             v.CategoryID,
		"permission_overwrites": ows,
	}
	var ch struct {
		ID string `json:"id"`
	}
	if err := v.call(ctx, client, http.MethodPost, "/guilds/"+v.GuildID+"/channels", body, &ch, nil); err != nil {
		return "", err
	}
	if ch.ID == "" {
		return "", errors.New("discord: channel response carried no id")
	}
	return ch.ID, nil
}

// GrantMember lets one server member into a room.
// PUT /channels/{id}/permissions/{user}.
func (v Voice) GrantMember(ctx context.Context, client *http.Client, channelID, userID string) error {
	body := overwrite{Type: overwriteMember, Allow: perms(PlayerAllow), Deny: perms(0)}
	return v.call(ctx, client, http.MethodPut, "/channels/"+channelID+"/permissions/"+userID, body, nil, nil)
}

// RevokeMember removes one user's overwrite, so the room is hidden from
// them again. It does not disconnect them; DisconnectFrom does. A room
// or an overwrite that is already gone is success.
func (v Voice) RevokeMember(ctx context.Context, client *http.Client, channelID, userID string) error {
	return v.call(ctx, client, http.MethodDelete, "/channels/"+channelID+"/permissions/"+userID, nil, nil,
		[]int{codeUnknownChannel, codeUnknownOverwrite})
}

// DeleteRoom deletes a room, which disconnects everyone in it and
// deletes its text chat. A room that is already gone is success.
func (v Voice) DeleteRoom(ctx context.Context, client *http.Client, channelID string) error {
	return v.call(ctx, client, http.MethodDelete, "/channels/"+channelID, nil, nil, []int{codeUnknownChannel})
}

// Rooms lists the voice channels in the category: the boot reconcile's
// view of what exists (ADR 0159 §5).
func (v Voice) Rooms(ctx context.Context, client *http.Client) ([]Room, error) {
	var chans []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     int    `json:"type"`
		ParentID string `json:"parent_id"`
	}
	if err := v.call(ctx, client, http.MethodGet, "/guilds/"+v.GuildID+"/channels", nil, &chans, nil); err != nil {
		return nil, err
	}
	var rooms []Room
	for _, c := range chans {
		if c.Type == channelTypeVoice && c.ParentID == v.CategoryID {
			rooms = append(rooms, Room{ID: c.ID, Name: c.Name})
		}
	}
	return rooms, nil
}

// IsMember reports whether a Discord user is in the community server.
// GET /guilds/{guild}/members/{user}, a REST call that needs no
// privileged gateway intent. Unknown member or user is false, not an
// error; an unknown guild is ErrVoiceNotFound.
func (v Voice) IsMember(ctx context.Context, client *http.Client, userID string) (bool, error) {
	err := v.call(ctx, client, http.MethodGet, "/guilds/"+v.GuildID+"/members/"+userID, nil, nil, nil)
	var nf *notFoundError
	if errors.As(err, &nf) && (nf.code == codeUnknownMember || nf.code == codeUnknownUser) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// VoiceState is a member's voice connection in the community server.
type VoiceState struct {
	// ChannelID is the channel they are connected to, "" for none.
	ChannelID string
	// Muted is true when they are muted or deafened, by themselves or by
	// the server.
	Muted bool
}

// VoiceStateOf is a member's voice connection in the community server.
// GET /guilds/{guild}/voice-states/{user}. A member in no channel is the
// zero VoiceState.
func (v Voice) VoiceStateOf(ctx context.Context, client *http.Client, userID string) (VoiceState, error) {
	var vs struct {
		ChannelID *string `json:"channel_id"`
		Deaf      bool    `json:"deaf"`
		Mute      bool    `json:"mute"`
		SelfDeaf  bool    `json:"self_deaf"`
		SelfMute  bool    `json:"self_mute"`
	}
	err := v.call(ctx, client, http.MethodGet, "/guilds/"+v.GuildID+"/voice-states/"+userID, nil, &vs, nil)
	var nf *notFoundError
	if errors.As(err, &nf) && (nf.code == codeUnknownVoiceState || nf.code == codeUnknownMember || nf.code == codeUnknownUser) {
		return VoiceState{}, nil
	}
	if err != nil {
		return VoiceState{}, err
	}
	if vs.ChannelID == nil {
		return VoiceState{}, nil
	}
	return VoiceState{ChannelID: *vs.ChannelID, Muted: vs.Deaf || vs.Mute || vs.SelfDeaf || vs.SelfMute}, nil
}

// ConnectedChannel is the voice channel a member is connected to in the
// community server, or "" when they are in none.
func (v Voice) ConnectedChannel(ctx context.Context, client *http.Client, userID string) (string, error) {
	st, err := v.VoiceStateOf(ctx, client, userID)
	return st.ChannelID, err
}

// MoveTo moves a member who is connected to some voice channel into
// channelID (PATCH /guilds/{guild}/members/{user}, Move Members). Discord
// refuses it for a member in no channel; the caller checks first.
func (v Voice) MoveTo(ctx context.Context, client *http.Client, channelID, userID string) error {
	body := map[string]any{"channel_id": channelID}
	return v.call(ctx, client, http.MethodPatch, "/guilds/"+v.GuildID+"/members/"+userID, body, nil,
		[]int{codeUnknownMember, codeUnknownUser})
}

// DisconnectFrom takes a member out of voice, but only if they are
// connected to channelID: a player who left a table is never pulled
// out of some other conversation. Reports whether it disconnected
// them. The check and the move are two calls, so a member who switches
// channels between them can be moved out of the new one; the window is
// a round trip.
func (v Voice) DisconnectFrom(ctx context.Context, client *http.Client, channelID, userID string) (bool, error) {
	current, err := v.ConnectedChannel(ctx, client, userID)
	if err != nil {
		return false, err
	}
	if current == "" || current != channelID {
		return false, nil
	}
	body := map[string]any{"channel_id": nil}
	err = v.call(ctx, client, http.MethodPatch, "/guilds/"+v.GuildID+"/members/"+userID, body, nil,
		[]int{codeUnknownMember, codeUnknownUser})
	if err != nil {
		return false, err
	}
	return true, nil
}

// notFoundError is a 404 with Discord's error code, for the callers
// that tell "unknown member" from "unknown guild". It unwraps to
// ErrVoiceNotFound.
type notFoundError struct{ code int }

func (e *notFoundError) Error() string {
	return fmt.Sprintf("%v (code %d)", ErrVoiceNotFound, e.code)
}
func (e *notFoundError) Unwrap() error { return ErrVoiceNotFound }

// voiceAPIError is the part of Discord's error body we read.
type voiceAPIError struct {
	Code       int     `json:"code"`
	RetryAfter float64 `json:"retry_after"`
}

// call sends one authenticated request. body may be nil; out may be nil
// to discard a 2xx body. A 404 whose Discord code is in goneOK is
// success: the thing being removed is already gone.
func (v Voice) call(ctx context.Context, client *http.Client, method, path string, body, out any, goneOK []int) error {
	if !v.Enabled() {
		return ErrVoiceNotConfigured
	}
	return botREST(ctx, client, v.Token, auditReason, method, path, body, out, goneOK)
}

// botREST is the one bot-token REST call the voice rooms and the LFG
// board (lfg.go) share. reason labels a change in the server's audit
// log. The errors are voice.go's sentinels; lfg.go renames the two
// that read differently for a message.
func botREST(ctx context.Context, client *http.Client, token, reason, method, path string, body, out any, goneOK []int) error {
	return botRESTLimit(ctx, client, token, reason, method, path, body, out, goneOK, maxDMBody)
}

// botRESTLimit is botREST with the cap on how much of the response it
// reads. The default (maxDMBody) suits the small objects every other
// call gets back; a listing, such as a guild's active threads
// (bugforum.go), needs more.
func botRESTLimit(ctx context.Context, client *http.Client, token, reason, method, path string, body, out any, goneOK []int, limit int64) error {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("discord: encode request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, botAPIBase+path, reader)
	if err != nil {
		return fmt.Errorf("discord: build request: %w", err)
	}
	// The only place the token appears. Never logged, never wrapped
	// into a returned error.
	req.Header.Set("Authorization", "Bot "+token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet {
		req.Header.Set("X-Audit-Log-Reason", reason)
	}

	resp, err := client.Do(req)
	if err != nil {
		// A transport error's text can carry the URL, never a header.
		return fmt.Errorf("%w: %v", ErrVoiceUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return fmt.Errorf("%w: read response: %v", ErrVoiceUnavailable, err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if out == nil || len(payload) == 0 {
			return nil
		}
		if err := json.Unmarshal(payload, out); err != nil {
			return fmt.Errorf("discord: parse response: %w", err)
		}
		return nil
	}

	var ae voiceAPIError
	_ = json.Unmarshal(payload, &ae)
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrBotUnauthorized
	case resp.StatusCode == http.StatusForbidden:
		return ErrVoiceForbidden
	case resp.StatusCode == http.StatusNotFound:
		for _, c := range goneOK {
			if ae.Code == c {
				return nil
			}
		}
		return &notFoundError{code: ae.Code}
	case resp.StatusCode == http.StatusTooManyRequests:
		wait := time.Duration(ae.RetryAfter * float64(time.Second))
		if wait <= 0 {
			if s, perr := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64); perr == nil && s > 0 {
				wait = time.Duration(s * float64(time.Second))
			}
		}
		if wait <= 0 {
			wait = time.Second
		}
		return &RateLimitError{RetryAfter: wait}
	case resp.StatusCode >= 500:
		return fmt.Errorf("%w: status %d", ErrVoiceUnavailable, resp.StatusCode)
	default:
		return fmt.Errorf("discord: bot call failed: status %d", resp.StatusCode)
	}
}
