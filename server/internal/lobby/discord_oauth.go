package lobby

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// discordConfig responds with a minimal flag so the Svelte Join
// screen can decide whether to show the "Sign in with Discord"
// button. Unauthenticated — a client must know this before having
// a session, by definition. Returns { enabled: true/false }
// and nothing else to keep secret details out of the response.
func discordConfig(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, map[string]bool{
		"enabled": c.Discord.Enabled(),
	})
}

// discordStart kicks off the OAuth round-trip. Expects
// ?game=<uuid>&t=<invite> on the URL; parks those in the state
// store and 302s the browser off to Discord's authorize page.
//
// Kept GET rather than POST because users reach this by clicking
// a plain link (either one the admin pasted into chat or the
// "Sign in with Discord" button on the Join page, which is just
// an <a href>). A POST would need a form + JS just to navigate.
func discordStart(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.Discord.Enabled() {
		return httpError(http.StatusServiceUnavailable, "Discord auth is not configured on this server")
	}
	gameStr := r.URL.Query().Get("game")
	invite := r.URL.Query().Get("t")
	if gameStr == "" || invite == "" {
		return httpError(http.StatusBadRequest, "game and t query params required")
	}
	gameID, err := uuid.Parse(gameStr)
	if err != nil {
		return httpError(http.StatusBadRequest, "game id must be a uuid")
	}

	store := c.discordStore()
	state, challenge, err := store.Start(gameID, invite)
	if err != nil {
		return fmt.Errorf("discord start: %w", err)
	}
	http.Redirect(w, r, c.Discord.AuthorizeURL(state, challenge), http.StatusFound)
	return nil
}

// discordCallback is Discord's bounce-back target. Validates the
// state, exchanges the code for a token, fetches the user's
// identity, claims the seat via JoinWithIdentity, mints a session,
// and redirects to the SPA with the session in the URL fragment.
//
// Fragment handoff: the token rides in `#...` rather than a query
// string so it never hits the server access log and isn't sent on
// any subsequent same-origin request's Referer header. The SPA's
// `oauth-complete` route parses the hash, persists the session,
// and navigates to the game route.
func discordCallback(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.Discord.Enabled() {
		return httpError(http.StatusServiceUnavailable, "Discord auth is not configured on this server")
	}

	// Discord can also bounce back with error/error_description
	// when the user denies consent. Surface as a plain 400 with
	// the reason so the SPA can show a clear message rather than
	// leaving the user in limbo.
	if derr := r.URL.Query().Get("error"); derr != "" {
		return httpError(http.StatusBadRequest,
			fmt.Sprintf("Discord auth failed: %s (%s)", derr, r.URL.Query().Get("error_description")))
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	if state == "" || code == "" {
		return httpError(http.StatusBadRequest, "state and code query params required")
	}

	entry, err := c.discordStore().Consume(state)
	if err != nil {
		return httpError(http.StatusBadRequest, "oauth state not found or expired")
	}

	client := c.DiscordHTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	token, err := c.Discord.ExchangeCode(r.Context(), client, code, entry.CodeVerifier)
	if err != nil {
		return httpError(http.StatusBadGateway, err.Error())
	}
	user, err := c.Discord.FetchUser(r.Context(), client, token.AccessToken)
	if err != nil {
		return httpError(http.StatusBadGateway, err.Error())
	}

	identity := DiscordIdentity{
		ID:         user.ID,
		Username:   user.Username,
		GlobalName: user.GlobalName,
		AvatarHash: user.Avatar,
	}
	meta, playerID, err := c.Lobby.JoinWithIdentity(entry.GameID, entry.InviteToken, "", identity)
	if err != nil {
		// Join failures (game full, game started, invalid invite, etc.)
		// surface as plain 4xx so the SPA's oauth-complete route can
		// show a meaningful error banner.
		return err
	}

	p := auth.Principal{
		Role:              auth.RolePlayer,
		GameID:            meta.ID,
		PlayerID:          playerID,
		Name:              identity.DisplayName(),
		DiscordID:         user.ID,
		DiscordUsername:   user.Username,
		DiscordGlobalName: user.GlobalName,
		DiscordAvatarHash: user.Avatar,
	}
	tok, issued, err := c.Auth.Issue(r.Context(), p, c.SessionTTL)
	if err != nil {
		return fmt.Errorf("issue session: %w", err)
	}
	setSessionCookie(c, w, tok, issued.ExpiresAt)

	// Hand the session off to the SPA via a URL fragment. The SPA
	// is served from the same origin, so a relative redirect is
	// enough — no need to hardcode cmd.labxp.io vs localhost.
	//
	// The fragment carries token, game, player_id, and expires_at
	// so the SPA can install the Session struct without another
	// round-trip to /me.
	frag := url.Values{}
	frag.Set("token", tok)
	frag.Set("game", meta.ID.String())
	frag.Set("player_id", playerID.String())
	frag.Set("expires_at", issued.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"))
	http.Redirect(w, r, "/#/oauth-complete?"+frag.Encode(), http.StatusFound)
	return nil
}

// discordAvatar serves a cached Discord avatar PNG. Path shape
// is /avatars/{id}/{hash}.png — the .png suffix is stripped here
// so the ServeMux pattern {hash} doesn't have to include it.
// Avatar cache handles disk I/O + CDN fetch + inflight dedup.
func discordAvatar(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.DiscordAvatars == nil {
		return httpError(http.StatusServiceUnavailable, "avatar cache not configured")
	}
	id := r.PathValue("id")
	hash := r.PathValue("hash")
	// The Svelte client builds the URL with `.png` on the end so
	// browsers pick the right decoder from the extension; the
	// cache keys don't carry the suffix, so strip it before
	// handing off.
	if len(hash) > 4 && hash[len(hash)-4:] == ".png" {
		hash = hash[:len(hash)-4]
	}
	err := c.DiscordAvatars.Serve(w, r, id, hash)
	if err == discord.ErrCacheDisabled {
		return httpError(http.StatusServiceUnavailable, "avatar cache disabled")
	}
	if err == discord.ErrInvalidAvatarKey {
		return httpError(http.StatusBadRequest, "invalid avatar id or hash")
	}
	if err != nil {
		return httpError(http.StatusBadGateway, err.Error())
	}
	return nil
}

// discordStore returns the Config's state store, constructing a
// fresh one on first use if the caller left the field nil. Stored
// back onto the Config so subsequent calls in the same request
// flow share a single map; this only matters in tests, since prod
// wires a single store at boot.
func (c *Config) discordStore() *discord.StateStore {
	if c.DiscordStateStore == nil {
		c.DiscordStateStore = discord.NewStateStore()
	}
	return c.DiscordStateStore
}
