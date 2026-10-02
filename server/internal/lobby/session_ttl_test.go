package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Tests for ADR 0110 Delivery PR 1, the sign-in fix: every session
// mint that carries a user inherits the expiry of the session it came
// from (or gets IdentityTTL when it came from a Discord sign-in), the
// sessions without a user keep SessionTTL, a signed-in spectator keeps
// their user, a signed-in seat can join the next table by code, a
// reclaim ticket carries the user only for the seat's own user, and
// Discord's prompt=none gets exactly one consent retry.

// minted decodes a session response with the wanted status.
func minted(t *testing.T, resp *http.Response, want int) sessionResponse {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d, want %d: %s", resp.StatusCode, want, body)
	}
	var out sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	if out.Token == "" {
		t.Fatal("response carries no token")
	}
	return out
}

// newPracticeRevocationStack is newRevocationStack with what POST
// /games/practice needs: the tutorial decks and a bot host.
func newPracticeRevocationStack(t *testing.T) revocationStack {
	t.Helper()
	return newRevocationStack(t, func(c *Config) {
		host := newFakeBotHost()
		c.Lobby.SetBotHost(host)
		c.Bots = host
		c.Cards = tutorialDeckIndex(t)
	})
}

// TestSignedInMintsInheritTheSourceExpiry is ADR 0110 §1 item 1, mint
// by mint: a signed-in caller's new session carries their user and
// expires at exactly the instant their sign-in does, not 12 hours from
// now and not 30 days from now.
func TestSignedInMintsInheritTheSourceExpiry(t *testing.T) {
	s := newPracticeRevocationStack(t)
	idTok := signInAs(t, s, "alice")
	id := mustValidate(t, s.auth, idTok)
	if id.UserID == uuid.Nil {
		t.Fatal("sign-in carries no user")
	}

	check := func(t *testing.T, what string, got sessionResponse, role auth.Role) auth.Principal {
		t.Helper()
		p := mustValidate(t, s.auth, got.Token)
		if p.Role != role || p.UserID != id.UserID {
			t.Errorf("%s: role %q user %s, want %q user %s", what, p.Role, p.UserID, role, id.UserID)
		}
		if !p.ExpiresAt.Equal(id.ExpiresAt) {
			t.Errorf("%s expires %v, want the sign-in's %v (%v from issue, not 12h)",
				what, p.ExpiresAt, id.ExpiresAt, p.ExpiresAt.Sub(p.IssuedAt))
		}
		if !got.ExpiresAt.Equal(p.ExpiresAt) {
			t.Errorf("%s: response says it expires %v, the token %v", what, got.ExpiresAt, p.ExpiresAt)
		}
		if p.DiscordID != id.DiscordID {
			t.Errorf("%s: Discord %q, want %q", what, p.DiscordID, id.DiscordID)
		}
		return p
	}

	// The invite link (POST /games/{id}/join).
	fnm, _ := s.lobby.Create("FNM")
	seat := check(t, "joinGame", minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", idTok,
		joinRequest{InviteToken: fnm.InviteToken}), http.StatusOK), auth.RolePlayer)

	// A pasted code (POST /join).
	cube, _ := s.lobby.Create("Cube")
	check(t, "joinByCode", minted(t, postJSON(t, s.srv, "/join", idTok,
		joinRequest{InviteToken: cube.InviteToken}), http.StatusOK), auth.RolePlayer)

	// Watching (POST /games/{id}/spectate).
	watch, _ := s.lobby.Create("Watch")
	spec := check(t, "spectateGame", minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
		spectateRequest{InviteToken: watch.SpectatorInvite, Name: "typed"}), http.StatusOK), auth.RoleSpectator)
	if spec.Name != "U" {
		t.Errorf("signed-in spectator is labelled %q, want the Discord display name", spec.Name)
	}

	// Back to a seat by user (POST /me/games/{id}/session).
	back := check(t, "myGameSession", minted(t, post(t, s.srv, "/me/games/"+fnm.ID.String()+"/session", idTok),
		http.StatusOK), auth.RolePlayer)
	if back.PlayerID != seat.PlayerID {
		t.Errorf("myGameSession seat %s, want %s", back.PlayerID, seat.PlayerID)
	}

	// The tutorial (POST /games/practice).
	check(t, "createPractice", minted(t, postJSON(t, s.srv, "/games/practice", idTok, nil),
		http.StatusCreated), auth.RolePlayer)

	// A reclaim ticket redeemed by the seat's own signed-in user.
	ticket := mintTicket(t, s.srv, adminToken(t, s.srv), fnm.ID, seat.PlayerID)
	check(t, "redeemSeatReclaim", minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/reclaim", idTok,
		reclaimRequest{Ticket: ticket.Ticket}), http.StatusOK), auth.RolePlayer)

	// The rule chains: a session minted from a seat session minted
	// from the sign-in still ends with the sign-in.
	seatTok := minted(t, post(t, s.srv, "/me/games/"+fnm.ID.String()+"/session", idTok), http.StatusOK).Token
	next, _ := s.lobby.Create("Next")
	check(t, "joinByCode from a seat session", minted(t, postJSON(t, s.srv, "/join", seatTok,
		joinRequest{InviteToken: next.InviteToken}), http.StatusOK), auth.RolePlayer)
	specTok := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
		spectateRequest{InviteToken: watch.SpectatorInvite}), http.StatusOK).Token
	later, _ := s.lobby.Create("Later")
	check(t, "joinGame from a spectator session", minted(t, postJSON(t, s.srv, "/games/"+later.ID.String()+"/join", specTok,
		joinRequest{InviteToken: later.InviteToken}), http.StatusOK), auth.RolePlayer)
}

// TestSessionsWithoutAUserKeepSessionTTL: guests and the admin token
// are unchanged, 12 hours from now, because nothing can revoke them.
func TestSessionsWithoutAUserKeepSessionTTL(t *testing.T) {
	s := newPracticeRevocationStack(t)
	lifetime := func(t *testing.T, what string, got sessionResponse) {
		t.Helper()
		p := mustValidate(t, s.auth, got.Token)
		if p.UserID != uuid.Nil {
			t.Errorf("%s carries user %s", what, p.UserID)
		}
		if d := p.ExpiresAt.Sub(p.IssuedAt); d != 12*time.Hour {
			t.Errorf("%s lives %v, want 12h", what, d)
		}
	}

	fnm, _ := s.lobby.Create("FNM")
	guest := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", "",
		joinRequest{InviteToken: fnm.InviteToken, Name: "Guest"}), http.StatusOK)
	lifetime(t, "guest joinGame", guest)

	cube, _ := s.lobby.Create("Cube")
	lifetime(t, "guest joinByCode", minted(t, postJSON(t, s.srv, "/join", "",
		joinRequest{InviteToken: cube.InviteToken, Name: "Guest"}), http.StatusOK))

	lifetime(t, "guest spectateGame", minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: fnm.SpectatorInvite, Name: "Watcher"}), http.StatusOK))

	lifetime(t, "guest createPractice", minted(t, postJSON(t, s.srv, "/games/practice", guest.Token, nil),
		http.StatusCreated))

	admin := adminToken(t, s.srv)
	if d := func() time.Duration { p := mustValidate(t, s.auth, admin); return p.ExpiresAt.Sub(p.IssuedAt) }(); d != 12*time.Hour {
		t.Errorf("admin session lives %v, want 12h", d)
	}

	// A reclaim ticket on its own is the whole credential: no user,
	// even for a signed-in person's seat.
	idTok := signInAs(t, s, "alice")
	seat := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", idTok,
		joinRequest{InviteToken: fnm.InviteToken}), http.StatusOK)
	ticket := mintTicket(t, s.srv, admin, fnm.ID, seat.PlayerID)
	lifetime(t, "redeemSeatReclaim with no session", redeem(t, s.srv, fnm.ID, ticket.Ticket, http.StatusOK))

	// Nor does someone else's sign-in make the ticket theirs.
	bob := signInAs(t, s, "bob")
	ticket = mintTicket(t, s.srv, admin, fnm.ID, seat.PlayerID)
	got := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/reclaim", bob,
		reclaimRequest{Ticket: ticket.Ticket}), http.StatusOK)
	lifetime(t, "redeemSeatReclaim with another user's session", got)
	if p := mustValidate(t, s.auth, got.Token); p.PlayerID != seat.PlayerID || p.DiscordID != "alice" {
		t.Errorf("ticket seated %s as %q, want alice's seat %s", p.PlayerID, p.DiscordID, seat.PlayerID)
	}
}

// TestSessionTTLSelection is the configured lifetimes: the three
// Discord callback mints (login page, invite link, link) get the full
// IdentityTTL, everything without a user gets SessionTTL.
func TestSessionTTLSelection(t *testing.T) {
	lifetime := func(t *testing.T, a auth.Authenticator, tok string) time.Duration {
		t.Helper()
		p := mustValidate(t, a, tok)
		return p.ExpiresAt.Sub(p.IssuedAt)
	}

	t.Run("defaults", func(t *testing.T) {
		s := newRevocationStack(t, nil)
		if got := lifetime(t, s.auth, signInAs(t, s, "alice")); got != 30*24*time.Hour {
			t.Errorf("identity session lives %v, want 720h", got)
		}
		if got := lifetime(t, s.auth, adminToken(t, s.srv)); got != 12*time.Hour {
			t.Errorf("admin session lives %v, want 12h", got)
		}
		if got := lifetime(t, s.auth, guestSeatToken(t, s.srv, s.lobby)); got != 12*time.Hour {
			t.Errorf("guest seat session lives %v, want 12h", got)
		}
	})

	t.Run("configured", func(t *testing.T) {
		s := newRevocationStack(t, func(c *Config) {
			c.SessionTTL = 2 * time.Hour
			c.IdentityTTL = 7 * 24 * time.Hour
		})
		if got := lifetime(t, s.auth, signInAs(t, s, "alice")); got != 7*24*time.Hour {
			t.Errorf("identity session lives %v, want 168h", got)
		}

		// The invite-link flow claims a seat in the callback itself.
		// It is a sign-in, so it gets the identity lifetime.
		meta, _ := s.lobby.Create("FNM")
		st, _, err := s.state.Start(meta.ID, meta.InviteToken)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+st+"&code=code-3"))
		if got := lifetime(t, s.auth, frag.Get("token")); got != 7*24*time.Hour {
			t.Errorf("invite-flow seat session lives %v, want 168h", got)
		}

		// Linking Discord to a guest seat is a sign-in too.
		guestMeta, _ := s.lobby.Create("Guests")
		guest := minted(t, postJSON(t, s.srv, "/games/"+guestMeta.ID.String()+"/join", "",
			joinRequest{InviteToken: guestMeta.InviteToken, Name: "Guest"}), http.StatusOK)
		if got := lifetime(t, s.auth, guest.Token); got != 2*time.Hour {
			t.Errorf("guest seat session lives %v, want 2h", got)
		}
		st, _, err = s.state.StartLink(guestMeta.ID, guest.PlayerID)
		if err != nil {
			t.Fatalf("StartLink: %v", err)
		}
		req, _ := http.NewRequest(http.MethodGet, s.srv.URL+"/auth/discord/callback?state="+st+"&code=code-4", nil)
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: guest.Token})
		client := *s.srv.Client()
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("link callback: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Fatalf("link callback: %d", resp.StatusCode)
		}
		linked := fragmentParams(t, resp.Header.Get("Location")).Get("token")
		if got := lifetime(t, s.auth, linked); got != 7*24*time.Hour {
			t.Errorf("linked seat session lives %v, want 168h", got)
		}
		if got := lifetime(t, s.auth, adminToken(t, s.srv)); got != 2*time.Hour {
			t.Errorf("admin session lives %v, want 2h", got)
		}
	})

	t.Run("no database", func(t *testing.T) {
		// No database, no user: the identity-only session keeps the
		// long lifetime ADR 0051 decision 3 gave it, and the seat it
		// claims is a session with no user, so SessionTTL.
		t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
		srv, l, _, state := newDiscordTestStack(t)
		idTok := identityTokenFromCallback(t, srv, state)
		meta, _ := l.Create("FNM")
		seat := minted(t, postJSON(t, srv, "/join", idTok, joinRequest{InviteToken: meta.InviteToken}), http.StatusOK)
		if seat.Principal.UserID != uuid.Nil {
			t.Fatalf("seat carries a user with no database: %s", seat.Principal.UserID)
		}
		if d := seat.ExpiresAt.Sub(seat.Principal.IssuedAt); d != 12*time.Hour {
			t.Errorf("no-database seat session lives %v, want 12h", d)
		}
		resp := doGet(t, srv, "/me", idTok)
		var me auth.Principal
		_ = json.NewDecoder(resp.Body).Decode(&me)
		resp.Body.Close()
		if d := me.ExpiresAt.Sub(me.IssuedAt); d != 30*24*time.Hour {
			t.Errorf("no-database identity session lives %v, want 720h", d)
		}
	})
}

// TestSignedInSeatJoinsTheNextTableByCode is ADR 0110 §1 item 3: once
// seat sessions last as long as the sign-in, a signed-in player at one
// table must be able to paste the code for the next. A guest seat and
// the admin token are still refused.
func TestSignedInSeatJoinsTheNextTableByCode(t *testing.T) {
	s := newRevocationStack(t, nil)
	idTok := signInAs(t, s, "alice")
	user := mustValidate(t, s.auth, idTok).UserID

	first, _ := s.lobby.Create("First")
	seat := minted(t, postJSON(t, s.srv, "/join", idTok, joinRequest{InviteToken: first.InviteToken}), http.StatusOK)

	second, _ := s.lobby.Create("Second")
	got := minted(t, postJSON(t, s.srv, "/join", seat.Token,
		joinRequest{InviteToken: second.InviteToken, Name: "Mallory"}), http.StatusOK)
	p := mustValidate(t, s.auth, got.Token)
	if p.GameID != second.ID || p.UserID != user || p.DiscordID != "alice" || p.Name != "U" {
		t.Errorf("second seat session = %+v, want alice's user at %s under her Discord name", p, second.ID)
	}
	meta, _ := s.lobby.Get(second.ID)
	if s, ok := findSeat(meta.Players, got.PlayerID); !ok || s.UserID != user.String() {
		t.Errorf("second seat is not alice's: %+v", s)
	}

	// A signed-in spectator can too.
	watch, _ := s.lobby.Create("Watch")
	spec := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
		spectateRequest{InviteToken: watch.SpectatorInvite}), http.StatusOK)
	third, _ := s.lobby.Create("Third")
	minted(t, postJSON(t, s.srv, "/join", spec.Token, joinRequest{InviteToken: third.InviteToken}), http.StatusOK)

	// Refused: a guest seat (no identity to carry over), a guest
	// spectator, and the admin token.
	fourth, _ := s.lobby.Create("Fourth")
	guestSpec := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: watch.SpectatorInvite, Name: "Watcher"}), http.StatusOK)
	for name, tok := range map[string]string{
		"guest seat":      guestSeatToken(t, s.srv, s.lobby),
		"guest spectator": guestSpec.Token,
		"admin":           adminToken(t, s.srv),
	} {
		resp := postJSON(t, s.srv, "/join", tok, joinRequest{InviteToken: fourth.InviteToken, Name: "X"})
		resp.Body.Close()
		if resp.StatusCode != http.StatusConflict {
			t.Errorf("%s joining by code: %d, want 409", name, resp.StatusCode)
		}
	}
}

// TestSignedInSpectatorKeepsTheirUser is ADR 0110 §1 item 2: watching a
// table no longer signs a person out. The spectator session carries the
// user and Discord identity, /me/* still answers, and the WebSocket
// binding carries the user so eviction reaches it.
func TestSignedInSpectatorKeepsTheirUser(t *testing.T) {
	s := newRevocationStack(t, nil)
	idTok := signInAs(t, s, "alice")
	user := mustValidate(t, s.auth, idTok).UserID
	watch, _ := s.lobby.Create("Watch")
	spec := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
		spectateRequest{InviteToken: watch.SpectatorInvite}), http.StatusOK)
	p := mustValidate(t, s.auth, spec.Token)
	if p.Role != auth.RoleSpectator || p.UserID != user || p.DiscordID != "alice" || p.DiscordAvatarHash != "av" {
		t.Fatalf("spectator session = %+v", p)
	}
	if p.PlayerID != uuid.Nil {
		t.Errorf("spectator holds a seat: %s", p.PlayerID)
	}

	for _, path := range []string{"/me/games", "/me/tablemates", "/me/decks"} {
		resp := doGet(t, s.srv, path, spec.Token)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s as a signed-in spectator: %d, want 200", path, resp.StatusCode)
		}
	}

	r := httptest.NewRequest(http.MethodGet, "/ws?game="+watch.ID.String(), nil)
	r.Header.Set("Authorization", "Bearer "+spec.Token)
	b, err := (&WSAuthorizer{Auth: s.auth}).AuthorizeUpgrade(r)
	if err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if b.UserID != user || !b.ReadOnly || b.PlayerID != uuid.Nil {
		t.Errorf("binding = %+v, want read-only with alice's user", b)
	}

	// A guest spectator is still nobody in particular.
	guest := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: watch.SpectatorInvite, Name: "Watcher"}), http.StatusOK)
	if gp := mustValidate(t, s.auth, guest.Token); gp.UserID != uuid.Nil || gp.DiscordID != "" || gp.Name != "Watcher" {
		t.Errorf("guest spectator = %+v", gp)
	}
	resp := doGet(t, s.srv, "/me/games", guest.Token)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("GET /me/games as a guest spectator: %d, want 403", resp.StatusCode)
	}
}

// TestRevocationStillKillsInheritedSessions: a long-lived session is
// only safe because it can be revoked (ADR 0051 decision 6). Every
// session minted from a sign-in carries the user, so signing out
// everywhere ends all of them, on HTTP and on the WebSocket.
func TestRevocationStillKillsInheritedSessions(t *testing.T) {
	s := newPracticeRevocationStack(t)
	idTok := signInAs(t, s, "alice")
	fnm, _ := s.lobby.Create("FNM")
	watch, _ := s.lobby.Create("Watch")
	cube, _ := s.lobby.Create("Cube")

	seat := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", idTok,
		joinRequest{InviteToken: fnm.InviteToken}), http.StatusOK)
	sessions := map[string]string{
		"sign-in":    idTok,
		"joinGame":   seat.Token,
		"joinByCode": minted(t, postJSON(t, s.srv, "/join", idTok, joinRequest{InviteToken: cube.InviteToken}), http.StatusOK).Token,
		"spectate": minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
			spectateRequest{InviteToken: watch.SpectatorInvite}), http.StatusOK).Token,
		"myGameSession": minted(t, post(t, s.srv, "/me/games/"+fnm.ID.String()+"/session", idTok), http.StatusOK).Token,
		"practice":      minted(t, postJSON(t, s.srv, "/games/practice", idTok, nil), http.StatusCreated).Token,
	}
	ticket := mintTicket(t, s.srv, adminToken(t, s.srv), fnm.ID, seat.PlayerID)
	sessions["reclaim"] = minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/reclaim", idTok,
		reclaimRequest{Ticket: ticket.Ticket}), http.StatusOK).Token

	resp := post(t, s.srv, "/logout/everywhere", sessions["spectate"])
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout-everywhere from a spectator session: %d, want 204", resp.StatusCode)
	}
	for name, tok := range sessions {
		if code, msg := statusOf(t, s.srv, tok); code != http.StatusUnauthorized || msg != "session revoked" {
			t.Errorf("%s after logout-everywhere: %d %q, want 401 \"session revoked\"", name, code, msg)
		}
	}
	r := httptest.NewRequest(http.MethodGet, "/ws?game="+watch.ID.String(), nil)
	r.Header.Set("Authorization", "Bearer "+sessions["spectate"])
	if _, err := (&WSAuthorizer{Auth: s.auth}).AuthorizeUpgrade(r); err == nil {
		t.Error("a revoked spectator session opened a WebSocket")
	}
}

// --- Discord prompt=none (ADR 0110 §2) ---------------------------------

// authorizeLocation GETs path and returns the 302's Location, parsed.
func authorizeLocation(t *testing.T, srv *httptest.Server, path string, cookie string) *url.URL {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: cookie})
	}
	client := *srv.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("GET %s: %d, want 302", path, resp.StatusCode)
	}
	u, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("parse Location: %v", err)
	}
	return u
}

func TestDiscordStartSendsPromptNone(t *testing.T) {
	srv, l, _, store := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	for _, path := range []string{
		"/auth/discord/start",
		"/auth/discord/start?game=" + meta.ID.String() + "&t=" + meta.InviteToken,
	} {
		u := authorizeLocation(t, srv, path, "")
		if got := u.Query().Get("prompt"); got != "none" {
			t.Errorf("%s: prompt=%q, want none", path, got)
		}
		if e, err := store.Consume(u.Query().Get("state")); err != nil || e.Prompt != discord.PromptNone {
			t.Errorf("%s: parked %+v (%v), want a prompt=none entry", path, e, err)
		}
	}

	// "Sign in with a different Discord account".
	u := authorizeLocation(t, srv, "/auth/discord/start?prompt=consent", "")
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Errorf("?prompt=consent: prompt=%q", got)
	}
	resp := doGet(t, srv, "/auth/discord/start?prompt=login", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("?prompt=login: %d, want 400", resp.StatusCode)
	}
}

// TestDiscordLinkAsksForConsent: attaching an account to a seat always
// shows Discord's screen, so the person sees which account it is.
func TestDiscordLinkAsksForConsent(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	guest := minted(t, postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Guest"}), http.StatusOK)
	u := authorizeLocation(t, srv, "/auth/discord/link?game="+meta.ID.String(), guest.Token)
	if got := u.Query().Get("prompt"); got != "consent" {
		t.Errorf("link: prompt=%q, want consent", got)
	}
}

// TestDiscordPromptNoneRefusalIsRetriedOnceWithConsent is §2 item 2:
// an error on a prompt=none round sends the browser back to Discord
// with prompt=consent, for the same flow; an error on that round is
// shown. The retry happens once.
func TestDiscordPromptNoneRefusalIsRetriedOnceWithConsent(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, stub, store := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")
	first := authorizeLocation(t, srv, "/auth/discord/start?game="+meta.ID.String()+"&t="+meta.InviteToken, "")
	st := first.Query().Get("state")

	retry := authorizeLocation(t, srv, "/auth/discord/callback?error=consent_required&state="+st, "")
	if !strings.Contains(retry.String(), "discord.com/oauth2/authorize") {
		t.Fatalf("retry goes to %s, want Discord's authorize page", retry)
	}
	q := retry.Query()
	if q.Get("prompt") != "consent" {
		t.Errorf("retry prompt=%q, want consent", q.Get("prompt"))
	}
	st2 := q.Get("state")
	if st2 == "" || st2 == st || q.Get("code_challenge") == first.Query().Get("code_challenge") {
		t.Error("the retry must have its own state and PKCE challenge")
	}
	if stub.tokenWasCalled() {
		t.Error("a refused round reached Discord's token endpoint")
	}

	// The refused round's state is spent.
	resp := doGet(t, srv, "/auth/discord/callback?error=consent_required&state="+st, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("replaying the refused state: %d, want 400", resp.StatusCode)
	}

	// An error on the consent round is a real refusal: shown, not
	// retried, whatever the error is.
	e, err := store.Consume(st2)
	if err != nil {
		t.Fatalf("retry state not parked: %v", err)
	}
	if e.Prompt != discord.PromptConsent || e.GameID != meta.ID || e.InviteToken != meta.InviteToken {
		t.Errorf("retry entry = %+v, want the same invite flow with consent", e)
	}
	st3, _, _ := store.StartWithPrompt(e.GameID, e.InviteToken, discord.PromptConsent)
	resp = doGet(t, srv, "/auth/discord/callback?error=consent_required&state="+st3, "")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "consent_required") {
		t.Errorf("an error on a consent round: %d %s, want 400 naming the error", resp.StatusCode, body)
	}

	// The retry finishes the flow it was retrying: the code on the
	// consent round claims the seat the invite named.
	st4, _, _ := store.Retry(discord.StateEntry{GameID: meta.ID, InviteToken: meta.InviteToken, Prompt: discord.PromptNone})
	frag := fragmentParams(t, followOneRedirect(t, srv, "/auth/discord/callback?state="+st4+"&code=code-1"))
	if frag.Get("game") != meta.ID.String() || frag.Get("player_id") == "" {
		t.Errorf("the retried invite flow did not seat the player: %v", frag)
	}
}

// TestDiscordCallbackErrorWithoutAStateIsShown: nothing to retry.
func TestDiscordCallbackErrorWithoutAStateIsShown(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)
	for _, q := range []string{"error=access_denied", "error=access_denied&state=forged"} {
		resp := doGet(t, srv, "/auth/discord/callback?"+q, "")
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: %d, want 400", q, resp.StatusCode)
		}
	}
}

// TestOnlyIssueForMintsSessions holds §1's one rule in place: every
// lobby route that mints a session goes through issueFor, so a new
// route cannot quietly hand a signed-in person a 12-hour session (or a
// guest a 30-day one) by calling the authenticator itself.
func TestOnlyIssueForMintsSessions(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	mint := regexp.MustCompile(`\.Auth\.(Issue|IssueUntil)\(`)
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") || f == "session_ttl.go" {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if mint.MatchString(line) {
				t.Errorf("%s:%d mints a session directly; call issueFor (session_ttl.go) instead", f, i+1)
			}
		}
	}
}
