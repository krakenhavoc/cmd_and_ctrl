package ws

// admin_seat_test.go — ADR 0110 §3 with the owner's requirement of
// 2026-10-02: an admin bound to their own seat plays as that seat but
// keeps the admin card overrides. The hub is what tells Dispatch the
// connection is an admin (actions.Action.Admin); this pins that it
// does, and that a plain seat connection still cannot.

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func tappedIn(g *game.Game, id uuid.UUID) bool {
	var tapped bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}

func TestSeatedAdminConnectionKeepsCardOverrides(t *testing.T) {
	for _, tc := range []struct {
		name  string
		admin bool
	}{
		{"admin seat", true},
		{"plain seat", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			_, g, room, cleanup := newE2EServerWithRoom(t)
			defer cleanup()

			owner, other := g.Seats[0], g.Seats[1]
			c := game.NewCard("Override Target", owner.ID)
			c.TypeLine = "Creature — Test"
			g.Battlefield.PushTop(c)

			hub := NewHub(log)
			hub.SetRoom(room)
			hub.SetAuthorizer(UpgradeAuthorizerFunc(func(r *http.Request) (Binding, error) {
				return Binding{GameID: g.ID, PlayerID: other.ID, Admin: tc.admin}, nil
			}))
			mux := http.NewServeMux()
			mux.HandleFunc("GET /ws", hub.ServeWS)
			srv := httptest.NewServer(mux)
			defer srv.Close()

			conn := dial(t, "ws"+strings.TrimPrefix(srv.URL, "http")+"/ws")
			defer conn.Close()
			readNextFrame(t, conn) // initial

			tap := protocol.ActionPayload{
				Type:   "tap",
				Params: json.RawMessage(`{"instance_id":"` + c.InstanceID.String() + `"}`),
			}
			if tc.admin {
				sendActionAndWait(t, conn, tap)
				if !tappedIn(g, c.InstanceID) {
					t.Error("an admin seated at the table could not tap another seat's card")
				}
				return
			}
			expectActionError(t, conn, tap)
			if tappedIn(g, c.InstanceID) {
				t.Error("a plain seat tapped another seat's card")
			}
		})
	}
}
