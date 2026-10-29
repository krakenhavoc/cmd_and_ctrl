package lobby

// Tests for ban_sync.go (ADR 0156, amendment 2026-10-29, #3325): the
// Discord bot reports a ban, a role taken away or a leave on the C2
// server, and the account's sessions end at once, once, with a line in
// the mod log; an account that is exempt, missing, or already out is
// left alone. Which gateway events call this route is the bot's half,
// tested in bot/bansync_test.go.

import (
	"encoding/json"
	"net/http"
	"testing"
)

const syncedMember = "333333333333333333"

// banSyncStack is the gated sign-up stack with a mod log, and the member
// signed in with the role.
type banSyncStack struct {
	signupStack
	fake *fakeModLog
	log  *ModLog
}

func newBanSyncStack(t *testing.T, gate bool, configure func(*Config)) banSyncStack {
	t.Helper()
	fake := &fakeModLog{}
	ml := newTestModLog(fake)
	s := newSignupStack(t, gate, func(c *Config) {
		c.ModLog = ml
		if configure != nil {
			configure(c)
		}
	})
	s.stub.setUser(`{"id":"` + syncedMember + `","username":"syncee","global_name":"Syncee","avatar":"av"}`)
	return banSyncStack{signupStack: s, fake: fake, log: ml}
}

// banSync calls the route as the admin token and returns the status and
// the outcome.
func (s banSyncStack) banSync(t *testing.T, discordID, reason string) (int, banSyncResponse) {
	t.Helper()
	resp := postJSON(t, s.srv, "/admin/discord-members/"+discordID+"/revoke-sessions", adminToken(t, s.srv),
		map[string]string{"reason": reason})
	defer resp.Body.Close()
	var out banSyncResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestBanSyncRevokesOnceAndLogs(t *testing.T) {
	s := newBanSyncStack(t, true, nil)
	tok := s.signInAdmitted(t)
	userID := mustValidate(t, s.auth, tok).UserID
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
		t.Fatalf("signed in: /me %d", code)
	}

	code, out := s.banSync(t, syncedMember, "banned")
	if code != http.StatusOK || out.Outcome != BanSyncRevoked || out.SocketsClosed != 1 {
		t.Fatalf("ban: %d %+v", code, out)
	}
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusUnauthorized {
		t.Errorf("the banned member's session still works: /me %d", code)
	}
	if _, res := s.access(t, userID); res.String != "not_member" {
		t.Errorf("access_result = %q, want not_member", res.String)
	}

	// The member-remove that comes with a ban finds them already out:
	// nothing more is revoked or logged.
	code, out = s.banSync(t, syncedMember, "left")
	if code != http.StatusOK || out.Outcome != BanSyncAlreadyOut {
		t.Fatalf("remove after the ban: %d %+v", code, out)
	}
	s.log.Wait()
	got := s.fake.posted()
	if len(got) != 1 {
		t.Fatalf("mod log entries = %+v, want one", got)
	}
	e := got[0]
	if e.Title != "Sessions revoked by ban sync" || field(e, "Account") != "Syncee" ||
		field(e, "Why") != "Banned from the C2 server" || field(e, "Connections closed") != "1" {
		t.Errorf("entry = %+v", e)
	}
	if e.Link != "https://cmd.example/#/admin/accounts/"+userID.String() {
		t.Errorf("link = %q", e.Link)
	}
	noIDs(t, e, userID.String(), syncedMember)
}

// A sign-in records "admitted", so after the role is given back and the
// person signs in, taking it away again revokes again.
func TestBanSyncActsAgainAfterTheNextSignIn(t *testing.T) {
	s := newBanSyncStack(t, true, nil)
	s.signInAdmitted(t)
	if _, out := s.banSync(t, syncedMember, "role_removed"); out.Outcome != BanSyncRevoked {
		t.Fatalf("first removal: %+v", out)
	}
	if _, out := s.banSync(t, syncedMember, "role_removed"); out.Outcome != BanSyncAlreadyOut {
		t.Fatalf("a repeated update: %+v", out)
	}

	afterAMillisecond()
	tok := s.signInAdmitted(t)
	userID := mustValidate(t, s.auth, tok).UserID
	if _, res := s.access(t, userID); res.String != "admitted" {
		t.Fatalf("access_result after signing in = %q", res.String)
	}
	if _, out := s.banSync(t, syncedMember, "role_removed"); out.Outcome != BanSyncRevoked {
		t.Fatalf("second removal: %+v", out)
	}
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusUnauthorized {
		t.Errorf("the new session survived: /me %d", code)
	}
	if _, res := s.access(t, userID); res.String != "no_role" {
		t.Errorf("access_result = %q, want no_role", res.String)
	}
	s.log.Wait()
	got := s.fake.posted()
	if len(got) != 2 || field(got[1], "Why") != "The player role was taken away" {
		t.Errorf("mod log entries = %+v", got)
	}
}

func TestBanSyncLeavesAloneWhoItMust(t *testing.T) {
	for _, tc := range []struct {
		name      string
		gate      bool
		configure func(*Config)
		signIn    bool
		want      string
	}{
		{"no account", true, nil, false, BanSyncNoAccount},
		{"on the gate's exemption list", true, func(c *Config) {
			c.Discord.Gate = c.Discord.Gate.WithExempt(syncedMember)
		}, true, BanSyncExempt},
		{"an allowlisted admin", true, func(c *Config) {
			c.Admins = NewAdmins(NewAdminList(syncedMember), nil)
		}, true, BanSyncExempt},
		{"the gate off", false, nil, true, BanSyncGateOff},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newBanSyncStack(t, tc.gate, tc.configure)
			var tok string
			if tc.signIn {
				tok = s.signInAdmitted(t)
			}
			for _, reason := range []string{"banned", "role_removed", "left"} {
				code, out := s.banSync(t, syncedMember, reason)
				if code != http.StatusOK || out.Outcome != tc.want {
					t.Fatalf("%s: %d %+v, want %s", reason, code, out, tc.want)
				}
			}
			if tok != "" {
				if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
					t.Errorf("the session was revoked: /me %d", code)
				}
			}
			s.evictor.mu.Lock()
			evicted := len(s.evictor.calls)
			s.evictor.mu.Unlock()
			if evicted != 0 {
				t.Error("sockets were closed")
			}
			s.log.Wait()
			if got := s.fake.posted(); len(got) != 0 {
				t.Errorf("mod log entries = %+v, want none", got)
			}
		})
	}
}

func TestBanSyncRefusesBadRequests(t *testing.T) {
	s := newBanSyncStack(t, true, nil)
	tok := s.signInAdmitted(t)

	for _, tc := range []struct{ id, reason string }{
		{"not-a-snowflake", "banned"},
		{syncedMember, "muted"},
		{syncedMember, ""},
	} {
		if code, _ := s.banSync(t, tc.id, tc.reason); code != http.StatusBadRequest {
			t.Errorf("%q %q: %d, want 400", tc.id, tc.reason, code)
		}
	}

	// A signed-in person who is not an admin may not call it.
	resp := postJSON(t, s.srv, "/admin/discord-members/"+syncedMember+"/revoke-sessions", tok,
		map[string]string{"reason": "banned"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("a player: %d, want 403", resp.StatusCode)
	}
	if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
		t.Errorf("a refused call revoked: /me %d", code)
	}
}

func TestMemberRemovalAccess(t *testing.T) {
	for reason, want := range map[MemberRemoval]string{
		MemberBanned: "not_member", MemberLeft: "not_member", MemberRoleRemoved: "no_role",
	} {
		got, ok := reason.access()
		if !ok || string(got) != want {
			t.Errorf("%s: %q %v, want %q", reason, got, ok, want)
		}
	}
	if _, ok := MemberRemoval("kicked").access(); ok {
		t.Error("an unknown reason was accepted")
	}
}
