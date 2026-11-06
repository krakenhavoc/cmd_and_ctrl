package discord

// linked_roles.go is the Discord half of ADR 0163, Linked Roles: the
// configuration, the application's role-connection metadata (registered
// with the bot token), the authorize URL of the role-connection grant,
// and the per-person PUT that writes their numbers with their own access
// token.
//
// Nothing here decides WHEN a person's numbers are pushed or what they
// are. That is internal/linkedroles; this file only says what the calls
// look like on the wire and sends them. Same rules as the rest of the
// package: a token appears in an Authorization header and nowhere else,
// and Discord's own error text is never forwarded.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// The two variables Linked Roles reads beside the sign-in config, the
// bot token and CMDCTRL_IDENTITY_KEY.
const (
	// AppIDEnv is the Discord application id, the same variable the
	// gateway bot reads.
	AppIDEnv = "CMDCTRL_DISCORD_APP_ID"
	// LinkedRolesEnv is the explicit switch: "1" turns Linked Roles on,
	// empty or "0" leaves it off.
	LinkedRolesEnv = "CMDCTRL_DISCORD_LINKED_ROLES"
)

// RoleConnectionPlatform is the platform_name a person's connection
// shows on their Discord profile (ADR 0163 §5).
const RoleConnectionPlatform = "CMD & CTRL"

// roleConnectionScopes is what the role-connection grant asks for
// (ADR 0163 §3): `identify`, to find the person's account by their
// Discord id, and `role_connections.write`, to write their numbers.
const roleConnectionScopes = "identify role_connections.write"

// RoleConnectionScope is the one scope a role-connection grant must
// carry for a push to work.
const RoleConnectionScope = "role_connections.write"

// The three metadata keys (ADR 0163 §2). They are what a role's
// requirement in Discord names, so they are never renamed.
const (
	MetaGamesPlayed = "games_played"
	MetaGamesWon    = "games_won"
	MetaJoined      = "joined"
)

// Discord's application role connection metadata types.
const (
	metaIntegerGreaterOrEqual = 2
	// metaDatetimeLessOrEqual is "the date is at least N days ago": the
	// value (an ISO 8601 date) is less than or equal to the server's
	// configured number of days before today.
	metaDatetimeLessOrEqual = 5
)

// RoleConnectionMetadata is one field of the application's metadata, as
// PUT /applications/{app}/role-connections/metadata takes it.
type RoleConnectionMetadata struct {
	Type        int    `json:"type"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// RoleConnectionFields are ADR 0163 §2's three fields, in the order a
// server owner sees them.
func RoleConnectionFields() []RoleConnectionMetadata {
	return []RoleConnectionMetadata{
		{Type: metaIntegerGreaterOrEqual, Key: MetaGamesPlayed, Name: "Games played",
			Description: "Finished CMD & CTRL games played"},
		{Type: metaIntegerGreaterOrEqual, Key: MetaGamesWon, Name: "Games won",
			Description: "Finished CMD & CTRL games won"},
		{Type: metaDatetimeLessOrEqual, Key: MetaJoined, Name: "Days since joining",
			Description: "Days since the CMD & CTRL account was made"},
	}
}

// RoleConnection is the body of PUT
// /users/@me/applications/{app}/role-connection. The zero value, sent
// with Empty, clears what the profile shows.
type RoleConnection struct {
	PlatformName     string
	PlatformUsername string
	GamesPlayed      int
	GamesWon         int
	Joined           time.Time
}

// MaxPlatformUsername is Discord's limit on platform_username.
const MaxPlatformUsername = 100

// body is the wire form. Metadata values are strings, as Discord asks.
func (rc RoleConnection) body() map[string]any {
	name := rc.PlatformUsername
	if r := []rune(name); len(r) > MaxPlatformUsername {
		name = string(r[:MaxPlatformUsername])
	}
	return map[string]any{
		"platform_name":     rc.PlatformName,
		"platform_username": name,
		"metadata": map[string]string{
			MetaGamesPlayed: strconv.Itoa(rc.GamesPlayed),
			MetaGamesWon:    strconv.Itoa(rc.GamesWon),
			MetaJoined:      rc.Joined.UTC().Format(time.RFC3339),
		},
	}
}

// emptyRoleConnection is what an account deletion sends: no platform
// name, no username and no metadata, so the profile shows nothing.
var emptyRoleConnection = map[string]any{
	"platform_name":     nil,
	"platform_username": nil,
	"metadata":          map[string]string{},
}

// LinkedRoles is the configuration. The zero value is "off", and every
// call on it refuses without touching the network.
type LinkedRoles struct {
	// AppID is CMDCTRL_DISCORD_APP_ID.
	AppID string
	// Token is the bot token (CMDCTRL_DISCORD_BOT_TOKEN), used only to
	// register the metadata.
	Token string
	// Requested is CMDCTRL_DISCORD_LINKED_ROLES=1.
	Requested bool
}

// Enabled reports whether the switch is on and the application id and
// the bot token are both set. The caller still needs the sign-in config,
// a database and CMDCTRL_IDENTITY_KEY before it turns the feature on.
func (lr LinkedRoles) Enabled() bool { return lr.Requested && lr.AppID != "" && lr.Token != "" }

// Missing names what a requested but not enabled config lacks, for the
// boot log. "" when nothing is missing or nothing was requested.
func (lr LinkedRoles) Missing() string {
	if !lr.Requested {
		return ""
	}
	var out []string
	if lr.AppID == "" {
		out = append(out, AppIDEnv)
	}
	if lr.Token == "" {
		out = append(out, BotTokenEnv)
	}
	return strings.Join(out, " and ")
}

// LinkedRolesFromEnv reads the switch, the application id and the bot
// token. The switch unset, empty or "0" is off with no error. Any other
// value than "1" is an error naming the variable, and so is an
// application id that is not a snowflake: either is a typo to fix, not
// a state to boot in. "1" without the id or the token is a LinkedRoles
// whose Enabled() is false and Missing() says why (the token is
// provisioned on production only).
func LinkedRolesFromEnv(getenv func(string) string) (LinkedRoles, error) {
	switch v := strings.TrimSpace(getenv(LinkedRolesEnv)); v {
	case "", "0":
		return LinkedRoles{}, nil
	case "1":
	default:
		return LinkedRoles{}, fmt.Errorf("%s must be 1 (on), 0 or empty (off)", LinkedRolesEnv)
	}
	app := strings.TrimSpace(getenv(AppIDEnv))
	if app != "" && !isSnowflake(app) {
		return LinkedRoles{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", AppIDEnv)
	}
	return LinkedRoles{AppID: app, Token: BotFromEnv(getenv).Token, Requested: true}, nil
}

// RegisterMetadata sends the three fields (ADR 0163 §2) with the bot
// token. Discord replaces the whole list, so sending it again is
// harmless, and every boot does.
func (lr LinkedRoles) RegisterMetadata(ctx context.Context, client *http.Client) error {
	if !lr.Enabled() {
		return ErrLinkedRolesOff
	}
	return botREST(ctx, client, lr.Token, "CMD & CTRL Linked Roles", http.MethodPut,
		"/applications/"+url.PathEscape(lr.AppID)+"/role-connections/metadata", RoleConnectionFields(), nil, nil)
}

// PutRoleConnection writes one person's connection with their own
// access token from the role-connection grant.
func (lr LinkedRoles) PutRoleConnection(ctx context.Context, client *http.Client, accessToken string, rc RoleConnection) error {
	return lr.putRoleConnection(ctx, client, accessToken, rc.body())
}

// ClearRoleConnection empties one person's connection, so their profile
// stops showing the numbers (ADR 0163 §7).
func (lr LinkedRoles) ClearRoleConnection(ctx context.Context, client *http.Client, accessToken string) error {
	return lr.putRoleConnection(ctx, client, accessToken, emptyRoleConnection)
}

// ErrLinkedRolesOff is a call on a config that is not enabled.
var ErrLinkedRolesOff = fmt.Errorf("discord: Linked Roles is off — %s=1, %s and %s must all be set", LinkedRolesEnv, AppIDEnv, BotTokenEnv)

// ErrRoleConnectionUnauthorized is a 401 on the PUT: the access token
// was refused, which after a fresh refresh means the person removed the
// app between the two calls.
var ErrRoleConnectionUnauthorized = errors.New("discord: the role-connection access token was refused")

func (lr LinkedRoles) putRoleConnection(ctx context.Context, client *http.Client, accessToken string, body any) error {
	if lr.AppID == "" {
		return ErrLinkedRolesOff
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("discord: encode role connection: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		botAPIBase+"/users/@me/applications/"+url.PathEscape(lr.AppID)+"/role-connection", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("discord: build role connection request: %w", err)
	}
	// The only place the person's access token appears.
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVoiceUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxDMBody))
	if err != nil {
		return fmt.Errorf("%w: read response: %v", ErrVoiceUnavailable, err)
	}
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode <= 299:
		return nil
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrRoleConnectionUnauthorized
	case resp.StatusCode == http.StatusTooManyRequests:
		var ae voiceAPIError
		_ = json.Unmarshal(payload, &ae)
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
		return fmt.Errorf("discord: role connection: status %d", resp.StatusCode)
	}
}

// RoleConnectionAuthorizeURL is the authorize URL of the role-connection
// grant (ADR 0163 §3): the scopes above, the same redirect URI as
// sign-in, and always the consent screen, so the person sees what they
// are agreeing to show.
func (c Config) RoleConnectionAuthorizeURL(state, codeChallenge string) string {
	return c.authorizeURL(state, codeChallenge, PromptConsent, roleConnectionScopes)
}

// HasScope reports whether a token response's space-separated scope
// string carries scope.
func (t TokenResponse) HasScope(scope string) bool {
	for _, s := range strings.Fields(t.Scope) {
		if s == scope {
			return true
		}
	}
	return false
}
