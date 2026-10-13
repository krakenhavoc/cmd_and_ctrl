package discord

// gate.go is ADR 0156 §1: sign-up gated by a role in the community
// server. The configuration (three variables, checked at boot), the
// member check a sign-in and a renewal both make, and the refresh-token
// exchange a renewal needs to make it.

import (
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

// The three variables (ADR 0156 §1 item 1).
const (
	// CommunityGuildEnv is the C2 server's snowflake.
	CommunityGuildEnv = "CMDCTRL_DISCORD_COMMUNITY_GUILD_ID"
	// PlayerRoleEnv is the role in that server that admits a person:
	// Elder Dragon Highlander on the C2 server, which a mod grants. A
	// member with only the role everyone gets on joining (Planeswalker)
	// is refused as a member without the role.
	PlayerRoleEnv = "CMDCTRL_DISCORD_PLAYER_ROLE_ID"
	// InviteURLEnv is the server's public invite link, shown by the
	// client through GET /community. Independent of the gate.
	InviteURLEnv = "CMDCTRL_DISCORD_INVITE_URL"
)

// memberEndpointBase is where GET /users/@me/guilds/{guild}/member
// lives. A package var, like the OAuth endpoints, so tests point it at
// an httptest.Server.
var memberEndpointBase = "https://discord.com/api/users/@me/guilds"

// maxRetryAfter bounds how long the member check waits on a 429 before
// its one retry. Discord's Retry-After is usually well under a second;
// a longer one is an outage as far as a person waiting on a redirect is
// concerned.
var maxRetryAfter = 2 * time.Second

// Gate is the sign-up gate: on when both ids are set, off when both are
// empty. GateFromEnv refuses one without the other.
type Gate struct {
	GuildID string
	RoleID  string
}

// Enabled reports whether the gate is on.
func (g Gate) Enabled() bool { return g.GuildID != "" && g.RoleID != "" }

// GateFromEnv reads the guild and the role. Both unset (or blank) is the
// gate off, a supported state and every dev loop's. One without the
// other is an error naming the missing variable, and a value that is not
// a Discord snowflake (digits only, at most 20) is an error naming its
// variable: either would otherwise leave a server that admits everyone,
// or no one, without saying so. The values are not secret, but the error
// does not echo them.
func GateFromEnv(getenv func(string) string) (Gate, error) {
	g := Gate{
		GuildID: strings.TrimSpace(getenv(CommunityGuildEnv)),
		RoleID:  strings.TrimSpace(getenv(PlayerRoleEnv)),
	}
	switch {
	case g.GuildID == "" && g.RoleID == "":
		return Gate{}, nil
	case g.RoleID == "":
		return Gate{}, fmt.Errorf("%s is set but %s is not: set both to turn the sign-up gate on, or neither to leave it off", CommunityGuildEnv, PlayerRoleEnv)
	case g.GuildID == "":
		return Gate{}, fmt.Errorf("%s is set but %s is not: set both to turn the sign-up gate on, or neither to leave it off", PlayerRoleEnv, CommunityGuildEnv)
	}
	for _, v := range []struct{ name, id string }{{CommunityGuildEnv, g.GuildID}, {PlayerRoleEnv, g.RoleID}} {
		if !isSnowflake(v.id) {
			return Gate{}, fmt.Errorf("%s is not a Discord id (digits only, at most 20)", v.name)
		}
	}
	return g, nil
}

// InviteURLFromEnv reads the community invite link. Unset is "" and
// hides every Discord placement. A set value must be an absolute https
// URL with a host: the client renders it as a link on public pages, so
// anything else (a typo, a javascript: URL) fails the boot rather than
// reaching a page.
func InviteURLFromEnv(getenv func(string) string) (string, error) {
	raw := strings.TrimSpace(getenv(InviteURLEnv))
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", fmt.Errorf("%s must be an https URL such as https://discord.gg/<code>", InviteURLEnv)
	}
	return u.String(), nil
}

func isSnowflake(s string) bool {
	if s == "" || len(s) > 20 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Access is the answer to "does this person hold the role?" (ADR 0156
// §1 item 3). The strings are what users.access_result stores.
type Access string

const (
	// AccessAdmitted: a member of the guild holding the role.
	AccessAdmitted Access = "admitted"
	// AccessNotMember: Discord answered 404, not a member of the guild.
	AccessNotMember Access = "not_member"
	// AccessNoRole: a member without the role.
	AccessNoRole Access = "no_role"
)

// ErrUnavailable wraps every member-check or refresh failure that is not
// an answer: a 5xx, a timeout, a 429 that persisted past its one retry,
// a body that would not parse. The gate calls this an outage; a sign-in
// fails closed on it and a renewal goes ahead (ADR 0156 §1 items 4, 5).
var ErrUnavailable = errors.New("discord: no usable answer")

// ErrRefreshRejected is Discord refusing a refresh token outright (a 400
// or 401 from the token endpoint, typically invalid_grant: the person
// de-authorized the app, or the token was already rotated). A renewal
// treats it as "no refresh token": not renewed, and the next sign-in is
// the check.
var ErrRefreshRejected = errors.New("discord: refresh token rejected")

// CheckMember asks Discord for the caller's own member record in the
// gate's guild, with the caller's access token, and reports whether it
// holds the gate's role. Only an answer is returned without an error:
// 200 (admitted or no role) and 404 (not a member). Anything else wraps
// ErrUnavailable. A 429 is retried once, after Discord's Retry-After
// (capped at maxRetryAfter).
//
// The member record (nickname, roles, join date) is read here and
// dropped; only the Access is kept.
func (c Config) CheckMember(ctx context.Context, client *http.Client, accessToken string) (Access, error) {
	if !c.Gate.Enabled() {
		return "", errors.New("discord: CheckMember with the gate off")
	}
	endpoint := memberEndpointBase + "/" + url.PathEscape(c.Gate.GuildID) + "/member"
	for attempt := 0; ; attempt++ {
		status, retryAfter, body, err := getMember(ctx, client, endpoint, accessToken)
		if err != nil {
			return "", fmt.Errorf("%w: member request: %v", ErrUnavailable, err)
		}
		switch {
		case status == http.StatusOK:
			var m struct {
				Roles []string `json:"roles"`
			}
			if err := json.Unmarshal(body, &m); err != nil {
				return "", fmt.Errorf("%w: parse member body: %v", ErrUnavailable, err)
			}
			for _, r := range m.Roles {
				if r == c.Gate.RoleID {
					return AccessAdmitted, nil
				}
			}
			return AccessNoRole, nil
		case status == http.StatusNotFound:
			return AccessNotMember, nil
		case status == http.StatusTooManyRequests && attempt == 0:
			t := time.NewTimer(retryAfter)
			select {
			case <-ctx.Done():
				t.Stop()
				return "", fmt.Errorf("%w: %v", ErrUnavailable, ctx.Err())
			case <-t.C:
			}
		default:
			return "", fmt.Errorf("%w: member check: status %d", ErrUnavailable, status)
		}
	}
}

// getMember is one GET of the member endpoint. retryAfter is Discord's
// Retry-After header, in seconds (it may be fractional), capped at
// maxRetryAfter; a missing or unreadable one is a second, capped too.
func getMember(ctx context.Context, client *http.Client, endpoint, accessToken string) (status int, retryAfter time.Duration, body []byte, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, 0, nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err = io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if err != nil {
		return 0, 0, nil, err
	}
	retryAfter = time.Second
	if secs, perr := strconv.ParseFloat(strings.TrimSpace(resp.Header.Get("Retry-After")), 64); perr == nil && secs >= 0 {
		retryAfter = time.Duration(secs * float64(time.Second))
	}
	if retryAfter > maxRetryAfter {
		retryAfter = maxRetryAfter
	}
	return resp.StatusCode, retryAfter, body, nil
}

// ExchangeRefreshToken trades a stored refresh token for a new access
// token (and, usually, a rotated refresh token, which the caller must
// store in place of the old one). A 400 or 401 from Discord wraps
// ErrRefreshRejected; every other failure wraps ErrUnavailable. Neither
// token is ever put in an error.
func (c Config) ExchangeRefreshToken(ctx context.Context, client *http.Client, refreshToken string) (TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.ClientID)
	form.Set("client_secret", c.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: build refresh request: %v", ErrUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: refresh request: %v", ErrUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("%w: read refresh body: %v", ErrUnavailable, err)
	}
	switch {
	case resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized:
		var errResp TokenErrorResponse
		_ = json.Unmarshal(body, &errResp)
		return TokenResponse{}, fmt.Errorf("%w: status %d %s", ErrRefreshRejected, resp.StatusCode, errResp.Error)
	case resp.StatusCode != http.StatusOK:
		return TokenResponse{}, fmt.Errorf("%w: refresh: status %d", ErrUnavailable, resp.StatusCode)
	}
	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil || tr.AccessToken == "" {
		return TokenResponse{}, fmt.Errorf("%w: refresh body has no access token", ErrUnavailable)
	}
	return tr, nil
}
