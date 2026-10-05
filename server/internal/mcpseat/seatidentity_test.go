package mcpseat

import (
	"context"
	"strings"
	"testing"
)

// #2273: with no display_name the seat sits under its MCP client's name.
func TestJoinDefaultsTheNameFromTheMCPClient(t *testing.T) {
	cases := []struct {
		client, given, want string
	}{
		{"claude-code", "", "Claude Code"},
		{"codex-mcp-client", "", "Codex"},
		{"codex", "", "Codex"},
		{"Some Client", "", "some-client"},
		{"", "", "Agent"},
		{"claude-code", "Ada", "Ada"},
	}
	for _, c := range cases {
		f := newFakeServer(t)
		s := newTestSeat(t, nil)
		s.SetClient(c.client)
		r, err := s.Join(context.Background(), JoinInput{InviteURL: f.inviteURL(), DisplayName: c.given})
		if err != nil || r.IsError {
			t.Fatalf("join(%q): %v %q", c.client, err, r.Text)
		}
		f.mu.Lock()
		got := f.joins[0].Name
		f.mu.Unlock()
		if got != c.want {
			t.Errorf("client %q, display_name %q: seat named %q, want %q", c.client, c.given, got, c.want)
		}
	}
}

// #2274: two binaries sharing a --state-dir at one table must not play one
// seat between them. The second is refused, naming the cause, and the
// first keeps its seat.
func TestASharedStateDirDoesNotShareASeat(t *testing.T) {
	f := newFakeServer(t)
	dir := t.TempDir() + "/state"
	s1 := newTestSeat(t, func(c *Config) { c.StateDir = dir })
	joinFake(t, f, s1)

	s2 := newTestSeat(t, func(c *Config) { c.StateDir = dir })
	r, err := s2.Join(context.Background(), JoinInput{InviteURL: f.inviteURL(), DisplayName: "Second"})
	if err != nil {
		t.Fatal(err)
	}
	if !r.IsError || !strings.Contains(r.Text, "--state-dir") || strings.Contains(r.Text, "resumed") {
		t.Fatalf("the second seat on a shared state dir was not refused:\n%s", r.Text)
	}
	s2.mu.Lock()
	holds := s2.sess != nil || s2.conn != nil
	s2.mu.Unlock()
	if holds {
		t.Error("the refused seat still took a session or a connection")
	}
	f.mu.Lock()
	joins := len(f.joins)
	f.mu.Unlock()
	if joins != 1 {
		t.Errorf("joins = %d; the refusal must come before any claim", joins)
	}

	// The claim is the first seat's for as long as it holds the seat,
	// and is released with it, so a restart still reattaches.
	s1.Close()
	r, _ = s2.Join(context.Background(), JoinInput{InviteURL: f.inviteURL()})
	if r.IsError || !strings.Contains(r.Text, "resumed: true") {
		t.Fatalf("after the first seat let go, the saved seat did not resume:\n%s", r.Text)
	}
}

// A different state dir (the documented way to run two agents) is not
// affected: both seats sit down, each with its own session file.
func TestSeparateStateDirsStillSeatTwoAgents(t *testing.T) {
	f := newFakeServer(t)
	a := newTestSeat(t, nil)
	b := newTestSeat(t, nil)
	joinFake(t, f, a)
	r, _ := b.Join(context.Background(), JoinInput{InviteURL: f.inviteURL()})
	if r.IsError {
		t.Fatalf("a seat with its own state dir was refused:\n%s", r.Text)
	}
}
