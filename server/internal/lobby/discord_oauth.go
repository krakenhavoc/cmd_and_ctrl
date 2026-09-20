package lobby

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
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

// discordStart kicks off the OAuth round-trip, in either of two
// shapes:
//
//   - ?game=<uuid>&t=<invite> — the invite-link flow. Both values
//     are parked in the state store and the callback claims that
//     seat directly.
//   - no query params — the login-page flow. Nothing to park; the
//     callback mints an identity-only session and the SPA collects
//     an invite code afterwards.
//
// One of the pair without the other is a 400 rather than a guess.
// Treating it as unbound would silently discard a seat claim the
// user asked for, and inventing the missing half is not possible.
//
// Kept GET rather than POST because users reach this by clicking
// a plain link (either one the admin pasted into chat or the
// "Sign in with Discord" button on the Join / Login page, which is
// just an <a href>). A POST would need a form + JS just to navigate.
func discordStart(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.Discord.Enabled() {
		return httpError(http.StatusServiceUnavailable, "Discord auth is not configured on this server")
	}
	gameStr := r.URL.Query().Get("game")
	invite := r.URL.Query().Get("t")
	if (gameStr == "") != (invite == "") {
		return httpError(http.StatusBadRequest, "game and t must be supplied together")
	}
	var gameID uuid.UUID
	if gameStr != "" {
		var err error
		gameID, err = uuid.Parse(gameStr)
		if err != nil {
			return httpError(http.StatusBadRequest, "game id must be a uuid")
		}
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

	// A link round-trip only finishes in the browser that started it:
	// the one whose session cookie still holds the seat being linked.
	// Checked before the code is exchanged, so a callback that fails it
	// costs nothing upstream. See seatSessionFromCookie.
	if entry.Link() {
		if _, err := seatSessionFromCookie(c, r, entry.GameID, entry.LinkPlayerID); err != nil {
			return err
		}
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

	// Record the person before anything else happens (ADR 0051
	// decision 2): the first sign-in mints a users row, a later one
	// refreshes its name and avatar. Done before the seat claim so a
	// failed write cannot leave a seat taken by someone who then gets
	// no session. The refresh token goes to the store and nowhere
	// else — it is sealed there, or discarded when no
	// CMDCTRL_IDENTITY_KEY is configured. With no database the store
	// is users.NoStore and u.ID is zero, which is today's session.
	scopes := token.Scope
	if scopes == "" {
		scopes = discord.RequestedScopes
	}
	u, err := c.userStore().UpsertFromDiscord(r.Context(), user, token.RefreshToken, scopes)
	if err != nil {
		// The store's error can name tables and columns; the browser
		// gets a plain sentence and the log gets the detail.
		c.logger().Error("record Discord sign-in failed", "discord_id", user.ID, "err", err)
		return httpError(http.StatusInternalServerError, "could not record your sign-in; try again")
	}

	// Seats this snowflake claimed before it had a users row — every
	// Discord seat imported from lobby/*.json, and any claimed while
	// the deployment had no database — become this user's now (ADR
	// 0051 "Migration" step 3). Idempotent, so it runs on every
	// sign-in and costs one indexed UPDATE when there is nothing left
	// to link. A failure is logged, not fatal: the seats keep their
	// pending id and the next sign-in tries again.
	if u.ID != uuid.Nil {
		if _, err := c.Lobby.LinkPendingSeats(user.ID, u.ID); err != nil {
			c.logger().Warn("linking pending seats failed", "discord_id", user.ID, "err", err)
		}
	}

	// Login-page flow: nobody named a table, so there is no seat to
	// claim. Mint an identity-only session and hand it to the SPA,
	// which shows the invite-code box. The fragment carries no game
	// or player_id — their absence is how the client tells the two
	// flows apart.
	if entry.Unbound() {
		p := auth.Principal{
			Role:              auth.RoleIdentified,
			UserID:            u.ID,
			Name:              identity.DisplayName(),
			DiscordID:         user.ID,
			DiscordUsername:   user.Username,
			DiscordGlobalName: user.GlobalName,
			DiscordAvatarHash: user.Avatar,
		}
		// The long-lived one (ADR 0051 decision 3): CMDCTRL_IDENTITY_TTL,
		// 30 days by default, where every other session gets
		// SessionTTL. Handler fills the default, so IdentityTTL is
		// never zero here.
		tok, issued, err := c.Auth.Issue(r.Context(), p, c.IdentityTTL)
		if err != nil {
			return fmt.Errorf("issue session: %w", err)
		}
		setSessionCookie(c, w, tok, issued.ExpiresAt)

		frag := url.Values{}
		frag.Set("token", tok)
		frag.Set("expires_at", issued.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"))
		frag.Set("name", identity.DisplayName())
		setUserIDFragment(frag, u.ID)
		http.Redirect(w, r, "/#/oauth-complete?"+frag.Encode(), http.StatusFound)
		return nil
	}

	if entry.Link() {
		return finishDiscordLink(c, w, r, entry, identity, user, u.ID)
	}

	meta, playerID, err := c.Lobby.JoinAs(entry.GameID, entry.InviteToken, "", identity, u.ID)
	if err != nil {
		// Join failures (game full, game started, invalid invite, etc.)
		// surface as plain 4xx so the SPA's oauth-complete route can
		// show a meaningful error banner.
		return err
	}

	p := auth.Principal{
		Role:              auth.RolePlayer,
		UserID:            u.ID,
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
	setUserIDFragment(frag, u.ID)
	http.Redirect(w, r, "/#/oauth-complete?"+frag.Encode(), http.StatusFound)
	return nil
}

// discordLink starts the link round-trip: GET /auth/discord/link, from
// a player session, attaches the caller's Discord account to the seat
// that session holds (ADR 0051 sub-PR 4, carried over from S12.5 #59).
// A guest seat becomes that person's seat; a Discord seat can be moved
// to a different account. Game state does not matter: this is how a
// guest at a live table signs in without leaving it.
//
// ?game=<uuid> is optional. When present it must be the session's own
// game. It is the client saying which table it thinks it is linking,
// and a session cookie for another table (a second tab that joined
// somewhere else) is refused here rather than linking the wrong seat.
//
// CSRF is the existing state mechanism plus one binding. The state
// parks (game, player). The callback then requires the browser's
// session COOKIE to hold that same seat before it touches anything.
// Without the binding, a link round-trip started by one person could be
// finished by another: send someone the Discord consent URL, and their
// account ends up on your seat, so your seat is labelled as them and
// their "My games" lists it. Cookie only, not the ?token= or
// Authorization fallbacks, because a navigation back from Discord
// carries the cookie and nothing else.
//
// A plain GET for the same reason /start is one: the client reaches it
// by navigating, and the answer is a 302 to Discord.
func discordLink(c Config, w http.ResponseWriter, r *http.Request) error {
	if !c.Discord.Enabled() {
		return httpError(http.StatusServiceUnavailable, "Discord auth is not configured on this server")
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if raw := r.URL.Query().Get("game"); raw != "" {
		want, err := uuid.Parse(raw)
		if err != nil {
			return httpError(http.StatusBadRequest, "game id must be a uuid")
		}
		if want != p.GameID {
			return httpError(http.StatusConflict,
				"this browser's session is for a different table; reopen the game you want to link and try again")
		}
	}
	// Refuse what the callback would refuse, before the consent
	// screen rather than after it.
	meta, err := c.Lobby.Get(p.GameID)
	if err != nil {
		return err
	}
	if meta.Archived() {
		return ErrGameArchived
	}
	seat, ok := findSeat(meta.Players, p.PlayerID)
	if !ok {
		return ErrPlayerNotInGame
	}
	if seat.IsBot {
		return ErrSeatIsBot
	}

	state, challenge, err := c.discordStore().StartLink(p.GameID, p.PlayerID)
	if err != nil {
		return fmt.Errorf("discord link start: %w", err)
	}
	http.Redirect(w, r, c.Discord.AuthorizeURL(state, challenge), http.StatusFound)
	return nil
}

// seatSessionFromCookie validates the request's session cookie and
// returns its principal if it is the player session for (gameID,
// playerID). Anything else is a 403: the link was started by a
// different session, or this browser has since moved to another seat.
func seatSessionFromCookie(c Config, r *http.Request, gameID, playerID uuid.UUID) (auth.Principal, error) {
	refuse := httpError(http.StatusForbidden,
		"this Discord link was started from a different session; open your game and choose Link Discord again")
	ck, err := r.Cookie(auth.SessionCookie)
	if err != nil || ck.Value == "" {
		return auth.Principal{}, refuse
	}
	p, err := c.Auth.Validate(r.Context(), ck.Value)
	if err != nil || p.Role != auth.RolePlayer || p.GameID != gameID || p.PlayerID != playerID {
		return auth.Principal{}, refuse
	}
	return p, nil
}

// finishDiscordLink is the link branch of the callback: the identity
// goes onto the seat, the table hears about it, and the browser gets a
// fresh player session for the same seat that now carries the user.
func finishDiscordLink(c Config, w http.ResponseWriter, r *http.Request, entry discord.StateEntry,
	identity DiscordIdentity, user discord.User, userID uuid.UUID) error {
	meta, seat, err := c.Lobby.LinkSeat(entry.GameID, entry.LinkPlayerID, identity, userID)
	if err != nil {
		return err
	}
	p := auth.Principal{
		Role:              auth.RolePlayer,
		UserID:            userID,
		GameID:            meta.ID,
		PlayerID:          seat.PlayerID,
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

	frag := url.Values{}
	frag.Set("token", tok)
	frag.Set("game", meta.ID.String())
	frag.Set("player_id", seat.PlayerID.String())
	frag.Set("expires_at", issued.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00"))
	frag.Set("name", identity.DisplayName())
	setUserIDFragment(frag, userID)
	http.Redirect(w, r, "/#/oauth-complete?"+frag.Encode(), http.StatusFound)
	return nil
}

// setUserIDFragment adds user_id to the oauth-complete fragment when the
// session carries one. The client builds its principal from the
// fragment rather than calling /me, and user_id is how it knows the
// session is revocable, so whether to offer "sign out everywhere"
// (ADR 0051 decision 6). Absent with no database.
func setUserIDFragment(frag url.Values, id uuid.UUID) {
	if id != uuid.Nil {
		frag.Set("user_id", id.String())
	}
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

// logger returns the Config's logger, or slog's default.
func (c Config) logger() *slog.Logger {
	if c.Log == nil {
		return slog.Default()
	}
	return c.Log
}

// userStore returns the Config's user store, or users.NoStore when none
// is wired (no database, and every test that does not opt in).
func (c Config) userStore() users.Store {
	if c.Users == nil {
		return users.NoStore{}
	}
	return c.Users
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
