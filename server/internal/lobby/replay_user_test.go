package lobby

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// TestReplayForASignedInPersonsOwnSeat pins ADR 0158 §1 item 3: the
// Tables page's History offers Replay on a finished game whatever table
// the person's session is bound to now, so a signed-in person may pull
// the replay of a game they sat at, filtered for their own seat, once it
// has ended. Nobody else gains anything.
func TestReplayForASignedInPersonsOwnSeat(t *testing.T) {
	f := newReplayFixture(t)
	aliceUser := uuid.New()
	f.lobby.mu.Lock()
	for i := range f.lobby.games[f.gameID].meta.Players {
		if f.lobby.games[f.gameID].meta.Players[i].PlayerID == f.alice {
			f.lobby.games[f.gameID].meta.Players[i].UserID = aliceUser.String()
		}
	}
	f.lobby.mu.Unlock()

	// Alice signed in, her session bound to no table, or another one.
	identified := f.session(t, auth.Principal{Role: auth.RoleIdentified, UserID: aliceUser, Name: "Alice"})
	elsewhere := f.session(t, auth.Principal{
		Role: auth.RolePlayer, UserID: aliceUser, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Alice",
	})
	stranger := f.session(t, auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New(), Name: "Eve"})

	for _, tok := range []string{identified, elsewhere} {
		if got := doAuthGet(t, f.srv, f.path(), tok); got.code != http.StatusForbidden {
			t.Errorf("in progress: status = %d, want 403", got.code)
		}
	}

	f.end(t)

	if got := doAuthGet(t, f.srv, f.path(), stranger); got.code != http.StatusForbidden {
		t.Errorf("a stranger: status = %d, want 403", got.code)
	}
	aliceSeat := seatIndex(t, f.view, f.alice)
	bobSeat := seatIndex(t, f.view, f.bob)
	for _, tok := range []string{identified, elsewhere} {
		got := doAuthGet(t, f.srv, f.path(), tok)
		if got.code != http.StatusOK {
			t.Fatalf("ended: status = %d, want 200 (%s)", got.code, snip(got.body))
		}
		for i, p := range decodeReplay(t, got.body) {
			if len(p.Game.Seats[aliceSeat].Hand.Cards) == 0 {
				t.Errorf("line %d: Alice must see her own hand", i)
			}
			if n := len(p.Game.Seats[bobSeat].Hand.Cards); n != 0 {
				t.Errorf("line %d: Bob's hand leaked to Alice: %d cards", i, n)
			}
		}
	}
}
