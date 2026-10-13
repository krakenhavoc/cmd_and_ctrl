package lobby

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/capacity"
)

// practice_guest_test.go covers POST /games/practice/guest (ADR 0156
// §2): a practice table for a visitor with no session at all.

func guestPractice(t *testing.T, do func(method, path, token, contentType, body string) *http.Response, token string) (*http.Response, sessionResponse) {
	t.Helper()
	resp := do(http.MethodPost, "/games/practice/guest", token, "application/json", "")
	var out sessionResponse
	if resp.StatusCode == http.StatusCreated {
		_ = json.NewDecoder(resp.Body).Decode(&out)
	}
	return resp, out
}

func TestGuestPracticeMintsAGuestAndSeatsIt(t *testing.T) {
	_, do, l, host := newPracticeServer(t)

	resp, out := guestPractice(t, do, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("no session: got %d, want 201", resp.StatusCode)
	}
	if out.Token == "" || out.Game == nil || !out.Game.Practice {
		t.Fatalf("response: %+v", out)
	}
	if c, ok := sessionCookie(resp); !ok || c.Value != out.Token {
		t.Error("the session cookie was not set to the guest's seat")
	}
	p := out.Principal
	if p.Role != auth.RolePlayer || p.UserID.String() != "00000000-0000-0000-0000-000000000000" ||
		p.Name != "Player" || p.GameID != out.Game.ID || p.PlayerID != out.PlayerID {
		t.Errorf("guest session: %+v", p)
	}
	if got := host.startedSeats(out.Game.ID); len(got) != 1 {
		t.Errorf("bot runner not started: %v", got)
	}

	// An ordinary guest: it cannot open a real table.
	resp2 := do(http.MethodPost, "/games", out.Token, "application/json", `{"name":"Mine"}`)
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Errorf("guest POST /games: got %d, want 403", resp2.StatusCode)
	}

	// One live table per guest session: asking again with the session
	// replaces the table rather than adding one.
	resp3, again := guestPractice(t, do, out.Token)
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusCreated || again.Game == nil || again.Game.ID == out.Game.ID {
		t.Fatalf("second call: %d %+v", resp3.StatusCode, again.Game)
	}
	if _, err := l.Get(out.Game.ID); err == nil {
		t.Error("the guest's first practice table is still open")
	}
	if _, err := l.Get(again.Game.ID); err != nil {
		t.Errorf("the replacement is missing: %v", err)
	}
}

func TestGuestPracticeIsRateLimitedPerIP(t *testing.T) {
	// Not newPracticeServer: it relaxes the limiters.
	l, host, _ := newPracticeLobby(t)
	srv := newTestServerWithConfig(t, Config{
		Lobby: l, Auth: newTestAuth(), AdminToken: "shared-admin-token",
		Cards: tutorialDeckIndex(t), Bots: host,
	})
	post := func() int {
		resp, err := srv.Client().Post(srv.URL+"/games/practice/guest", "application/json", strings.NewReader(""))
		if err != nil {
			t.Fatalf("post: %v", err)
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}
	for i := 0; i < 3; i++ {
		if got := post(); got != http.StatusCreated {
			t.Fatalf("call %d: got %d, want 201", i+1, got)
		}
	}
	if got := post(); got != http.StatusTooManyRequests {
		t.Errorf("fourth call within the window: got %d, want 429", got)
	}
}

func TestGuestPracticeIsHeldToTheCapacityCaps(t *testing.T) {
	_, do, l, _ := newPracticeServer(t)
	// One practice table costs more than a load of 1 allows.
	l.SetCapacity(capLimits(1), nil)

	resp := do(http.MethodPost, "/games/practice/guest", "", "application/json", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("got %d, want 503", resp.StatusCode)
	}
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["code"] != capacity.CodeServerFull || body["error"] != capacity.Message {
		t.Errorf("body: %v", body)
	}
	if resp.Header.Get("Retry-After") == "" {
		t.Error("no Retry-After on the cap")
	}
}
