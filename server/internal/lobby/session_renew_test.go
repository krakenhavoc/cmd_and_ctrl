package lobby

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// Tests for ADR 0110 Delivery PR 2: POST /me/session, renewal on use
// (§1 item 5, owner answer 1) and the reinstall of a saved identity
// (§1 item 6).

// testClock is a Config.Now the test can move forward.
type testClock struct {
	mu     sync.Mutex
	offset time.Duration
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return time.Now().Add(c.offset)
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offset += d
}

// newRenewStack is newRevocationStack with a clock the test controls.
func newRenewStack(t *testing.T, configure func(*Config)) (revocationStack, *testClock) {
	t.Helper()
	clock := &testClock{}
	s := newRevocationStack(t, func(c *Config) {
		c.Now = clock.Now
		if configure != nil {
			configure(c)
		}
	})
	return s, clock
}

// renew POSTs /me/session with tok as the bearer.
func renew(t *testing.T, s revocationStack, tok string) *http.Response {
	t.Helper()
	return post(t, s.srv, "/me/session", tok)
}

// cookieValue returns the value of the session cookie a response set.
func cookieValue(resp *http.Response) (string, bool) {
	c, ok := sessionCookie(resp)
	if !ok {
		return "", false
	}
	return c.Value, true
}

// samePrincipal compares two principals ignoring their two timestamps.
func samePrincipal(a, b auth.Principal) bool {
	a.IssuedAt, a.ExpiresAt, b.IssuedAt, b.ExpiresAt = time.Time{}, time.Time{}, time.Time{}, time.Time{}
	return a == b
}

// TestRenewSessionPastHalfLife: a signed-in session more than half
// spent comes back as a new token for a fresh IdentityTTL, and the
// cookie is set to it.
func TestRenewSessionPastHalfLife(t *testing.T) {
	s, clock := newRenewStack(t, nil)
	idTok := signInAs(t, s, "alice")
	old := mustValidate(t, s.auth, idTok)

	clock.advance(15*24*time.Hour + time.Minute)
	afterAMillisecond()
	resp := renew(t, s, idTok)
	cookie, set := cookieValue(resp)
	got := minted(t, resp, http.StatusOK)

	if got.Token == idTok {
		t.Fatal("a session past half its life came back unrenewed")
	}
	if !set || cookie != got.Token {
		t.Errorf("cookie = %q (set %v), want the renewed token", cookie, set)
	}
	p := mustValidate(t, s.auth, got.Token)
	if d := p.ExpiresAt.Sub(p.IssuedAt); d != 30*24*time.Hour {
		t.Errorf("renewed session lives %v, want a fresh 720h", d)
	}
	if !p.IssuedAt.After(old.IssuedAt) || !p.ExpiresAt.After(old.ExpiresAt) {
		t.Errorf("renewed %v..%v, old %v..%v: want later on both ends", p.IssuedAt, p.ExpiresAt, old.IssuedAt, old.ExpiresAt)
	}
	if !samePrincipal(p, old) {
		t.Errorf("renewal changed the principal:\n got  %+v\n want %+v", p, old)
	}
	if !got.ExpiresAt.Equal(p.ExpiresAt) || !samePrincipal(got.Principal, p) {
		t.Errorf("response %+v does not describe the token %+v", got, p)
	}
	// The old token is not revoked: another tab may still hold it, and
	// it runs out at its own expiry.
	if code, _ := statusOf(t, s.srv, idTok); code != http.StatusOK {
		t.Errorf("the old token after a renewal: %d, want 200", code)
	}
}

// TestRenewSessionBelowHalfLifeIsANoOp: before half-life the route
// reinstalls the presented session as it is. Same token, same expiry,
// and the cookie is set to it.
func TestRenewSessionBelowHalfLifeIsANoOp(t *testing.T) {
	s, clock := newRenewStack(t, nil)
	idTok := signInAs(t, s, "alice")
	old := mustValidate(t, s.auth, idTok)

	for _, at := range []time.Duration{0, 15*24*time.Hour - time.Minute} {
		clock.advance(at)
		resp := renew(t, s, idTok)
		cookie, set := cookieValue(resp)
		got := minted(t, resp, http.StatusOK)
		if got.Token != idTok {
			t.Errorf("%v into a 30-day session: a new token, want the same one", at)
		}
		if !got.ExpiresAt.Equal(old.ExpiresAt) {
			t.Errorf("%v in: expires %v, want %v", at, got.ExpiresAt, old.ExpiresAt)
		}
		if !set || cookie != idTok {
			t.Errorf("%v in: cookie = %q (set %v), want the presented token", at, cookie, set)
		}
	}
}

// TestRenewedSeatAndSpectatorKeepTheirPrincipal: renewal copies the
// whole principal. A seat session is still that seat at that table,
// a spectator session still watches that table, both still carry the
// user and the Discord identity; only the two timestamps move.
func TestRenewedSeatAndSpectatorKeepTheirPrincipal(t *testing.T) {
	s, clock := newRenewStack(t, nil)
	idTok := signInAs(t, s, "alice")
	fnm, _ := s.lobby.Create("FNM")
	watch, _ := s.lobby.Create("Watch")
	seat := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", idTok,
		joinRequest{InviteToken: fnm.InviteToken}), http.StatusOK)
	spec := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", idTok,
		spectateRequest{InviteToken: watch.SpectatorInvite}), http.StatusOK)

	clock.advance(16 * 24 * time.Hour)
	afterAMillisecond()
	for name, tc := range map[string]struct {
		tok  string
		role auth.Role
		game uuid.UUID
	}{
		"seat":      {seat.Token, auth.RolePlayer, fnm.ID},
		"spectator": {spec.Token, auth.RoleSpectator, watch.ID},
	} {
		old := mustValidate(t, s.auth, tc.tok)
		got := minted(t, renew(t, s, tc.tok), http.StatusOK)
		if got.Token == tc.tok {
			t.Errorf("%s: not renewed", name)
			continue
		}
		p := mustValidate(t, s.auth, got.Token)
		if !samePrincipal(p, old) {
			t.Errorf("%s: renewal changed the principal:\n got  %+v\n want %+v", name, p, old)
		}
		if p.Role != tc.role || p.GameID != tc.game || p.UserID != old.UserID || p.UserID == uuid.Nil {
			t.Errorf("%s: renewed as %+v", name, p)
		}
		if d := p.ExpiresAt.Sub(p.IssuedAt); d != 30*24*time.Hour {
			t.Errorf("%s: renewed session lives %v, want 720h", name, d)
		}
		if got.PlayerID != p.PlayerID {
			t.Errorf("%s: response player_id %s, token %s", name, got.PlayerID, p.PlayerID)
		}
	}
	if p := mustValidate(t, s.auth, seat.Token); p.PlayerID != seat.PlayerID {
		t.Fatalf("seat token names %s, want %s", p.PlayerID, seat.PlayerID)
	}
}

// TestRenewSessionRefusesASessionWithoutAUser: a session that cannot be
// revoked is never renewed or reinstalled. Each refused caller gets the
// status that tells the client what it is: 401 for no credential at
// all, 403 for a credential that is not a person.
func TestRenewSessionRefusesASessionWithoutAUser(t *testing.T) {
	s, clock := newRenewStack(t, nil)
	watch, _ := s.lobby.Create("Watch")
	guestSpec := minted(t, postJSON(t, s.srv, "/games/"+watch.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: watch.SpectatorInvite, Name: "Watcher"}), http.StatusOK)

	// A reclaim ticket on its own: a signed-in person's seat, but no user.
	idTok := signInAs(t, s, "alice")
	fnm, _ := s.lobby.Create("FNM")
	seat := minted(t, postJSON(t, s.srv, "/games/"+fnm.ID.String()+"/join", idTok,
		joinRequest{InviteToken: fnm.InviteToken}), http.StatusOK)
	admin := adminToken(t, s.srv)
	ticket := redeem(t, s.srv, fnm.ID, mintTicket(t, s.srv, admin, fnm.ID, seat.PlayerID).Ticket, http.StatusOK)

	clock.advance(11 * time.Hour) // past half of a 12-hour session
	for name, tok := range map[string]string{
		"admin token":     admin,
		"guest seat":      guestSeatToken(t, s.srv, s.lobby),
		"guest spectator": guestSpec.Token,
		"reclaim ticket":  ticket.Token,
	} {
		resp := renew(t, s, tok)
		_, set := cookieValue(resp)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: %d, want 403", name, resp.StatusCode)
		}
		if set {
			t.Errorf("%s: the refusal set a cookie", name)
		}
	}

	resp := renew(t, s, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no credential: %d, want 401", resp.StatusCode)
	}
	resp = renew(t, s, "v1.forged.token")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("a forged token: %d, want 401", resp.StatusCode)
	}
}

// TestRenewSessionIgnoresAQueryToken: the route sets the cookie from
// the credential it reads, so it never reads ?token=, which a
// cross-site form could use to plant its own session in this site.
func TestRenewSessionIgnoresAQueryToken(t *testing.T) {
	s, _ := newRenewStack(t, nil)
	idTok := signInAs(t, s, "alice")
	resp := post(t, s.srv, "/me/session?token="+idTok, "")
	_, set := cookieValue(resp)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized || set {
		t.Errorf("?token= only: %d (cookie set %v), want 401 and no cookie", resp.StatusCode, set)
	}
}

// TestRenewSessionReadsTheCookieFirst: the reinstall sends the saved
// identity as a bearer once the session that replaced it has expired,
// which is when the browser has dropped that cookie. While the cookie
// is still there it is the credential, as on every other route.
func TestRenewSessionReadsTheCookieFirst(t *testing.T) {
	s, _ := newRenewStack(t, nil)
	idTok := signInAs(t, s, "alice")
	admin := adminToken(t, s.srv)

	req, _ := http.NewRequest(http.MethodPost, s.srv.URL+"/me/session", nil)
	req.Header.Set("Authorization", "Bearer "+idTok)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookie, Value: admin})
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("admin cookie, identity bearer: %d, want 403 (the cookie is the credential)", resp.StatusCode)
	}

	// No cookie: the saved identity's bearer is reinstalled.
	resp = renew(t, s, idTok)
	cookie, set := cookieValue(resp)
	got := minted(t, resp, http.StatusOK)
	if got.Token != idTok || !set || cookie != idTok {
		t.Errorf("reinstall: token %q cookie %q (set %v), want the saved identity's", got.Token, cookie, set)
	}
}

// TestRenewSessionRespectsRevocation: a token issued before the user
// signed out everywhere is refused, renewed or not, and no cookie is
// set.
func TestRenewSessionRespectsRevocation(t *testing.T) {
	s, clock := newRenewStack(t, nil)
	laptop := signInAs(t, s, "alice")
	phone := signInAs(t, s, "alice")
	resp := post(t, s.srv, "/logout/everywhere", phone)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout everywhere: %d", resp.StatusCode)
	}
	afterAMillisecond()

	for _, at := range []time.Duration{0, 20 * 24 * time.Hour} {
		clock.advance(at)
		resp := renew(t, s, laptop)
		_, set := cookieValue(resp)
		var body struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized || body.Error != "session revoked" {
			t.Errorf("%v in: %d %q, want 401 \"session revoked\"", at, resp.StatusCode, body.Error)
		}
		if set {
			t.Errorf("%v in: a revoked session set a cookie", at)
		}
	}
}

// revokeDuringIssue revokes the user's sessions just before it mints,
// which is the window between the route's Validate and its mint.
type revokeDuringIssue struct {
	auth.Authenticator
	revoker SessionRevoker
	armed   atomic.Bool
}

func (a *revokeDuringIssue) Issue(ctx context.Context, p auth.Principal, ttl time.Duration) (string, auth.Principal, error) {
	if a.armed.Load() && p.UserID != uuid.Nil {
		if _, err := a.revoker.RevokeAll(ctx, p.UserID); err != nil {
			return "", auth.Principal{}, err
		}
		time.Sleep(2 * time.Millisecond)
	}
	return a.Authenticator.Issue(ctx, p, ttl)
}

// TestRenewSessionRevokedMidRenewalIsRefused: a revocation that lands
// after the route validated the old token and before it minted the new
// one does not let the new one out. Its IssuedAt is after the
// revocation, so nothing else would refuse it.
func TestRenewSessionRevokedMidRenewalIsRefused(t *testing.T) {
	var wrap *revokeDuringIssue
	s, clock := newRenewStack(t, func(c *Config) {
		wrap = &revokeDuringIssue{Authenticator: c.Auth, revoker: c.Revocations}
		c.Auth = wrap
	})
	idTok := signInAs(t, s, "alice")
	clock.advance(16 * 24 * time.Hour)
	wrap.armed.Store(true)

	resp := renew(t, s, idTok)
	_, set := cookieValue(resp)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized || set {
		t.Errorf("revoked mid-renewal: %d (cookie set %v), want 401 and no cookie", resp.StatusCode, set)
	}
}

// TestRenewSessionIsRateLimitedPerPerson: a bucket per user, so one
// person's tabs cannot spend another's allowance through a shared
// proxy address.
func TestRenewSessionIsRateLimitedPerPerson(t *testing.T) {
	s, _ := newRenewStack(t, func(*Config) {
		// newRevocationStack relaxes every limiter; this test needs
		// them real. The Handler reads the variable when it is built,
		// which is after this hook.
		t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "")
	})
	alice := signInAs(t, s, "alice")
	bob := signInAs(t, s, "bob")

	limited := false
	for i := 0; i < 20; i++ {
		resp := renew(t, s, alice)
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			limited = true
			if i < 10 {
				t.Errorf("limited after %d calls, want a burst of 10", i)
			}
			break
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("call %d: %d", i, resp.StatusCode)
		}
	}
	if !limited {
		t.Fatal("20 renewals in a row were never limited")
	}
	resp := renew(t, s, bob)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("another person after alice is limited: %d, want 200", resp.StatusCode)
	}
}
