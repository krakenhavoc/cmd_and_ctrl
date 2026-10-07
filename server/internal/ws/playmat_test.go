package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0128: a seat's playmat is stamped on every captured view, is
// absent for a seat nobody set one on, is never put on a bot or agent
// seat, and is the same for every viewer.
func TestRoomStampsPlaymatsOnHumanSeatsOnly(t *testing.T) {
	room, human, bot, agent := seatKindsRoom(t)
	other, err := room.Game.AddPlayer("Other", []game.Card{{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}})
	if err != nil {
		t.Fatal(err)
	}
	withNone := other.ID
	room.SetPlaymat(human, "/playmats/aaaa")
	room.SetPlaymat(bot, "/playmats/bbbb")   // must never reach the wire
	room.SetPlaymat(agent, "/playmats/cccc") // nor this one
	room.SetPlaymat(uuid.Nil, "/playmats/zzzz")

	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	want := map[uuid.UUID]string{human: "/playmats/aaaa", withNone: "", bot: "", agent: ""}
	for _, s := range view.Seats {
		id := uuid.MustParse(s.ID)
		if s.PlaymatURL != want[id] {
			t.Errorf("seat %s (bot=%v agent=%v): playmat %q, want %q", s.Name, s.IsBot, s.IsAgent, s.PlaymatURL, want[id])
		}
	}

	// Every viewer sees the same URL on every seat: a seat, a rival,
	// the spectator and the admin.
	viewers := []string{human.String(), withNone.String(), protocol.SpectatorViewerID, ""}
	for _, v := range viewers {
		f := protocol.FilterViewFor(view, v)
		for _, s := range f.Seats {
			if got := want[uuid.MustParse(s.ID)]; s.PlaymatURL != got {
				t.Errorf("viewer %q sees %q on seat %s, want %q", v, s.PlaymatURL, s.Name, got)
			}
		}
	}
}

func TestRoomClearsAndReplacesAPlaymat(t *testing.T) {
	g := seedTestGame(t)
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	p := g.Seats[0].ID
	seen := func() string {
		view, _, err := room.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		return view.Seats[0].PlaymatURL
	}
	room.SetPlaymat(p, "/playmats/one")
	if got := seen(); got != "/playmats/one" {
		t.Fatalf("got %q", got)
	}
	room.SetPlaymat(p, "/playmats/two")
	if got := seen(); got != "/playmats/two" {
		t.Errorf("after a replacement: %q", got)
	}
	room.SetPlaymat(p, "")
	if got := seen(); got != "" {
		t.Errorf("after clearing: %q", got)
	}
}

// ADR 0128 amendment: the owner's wash rides beside the URL, on the
// same seats, and never without one.
func TestRoomStampsTheWashOnlyBesideAPlaymat(t *testing.T) {
	room, human, bot, _ := seatKindsRoom(t)
	room.SetPlaymatWash(human, 70)
	room.SetPlaymat(bot, "/playmats/bbbb")
	room.SetPlaymatWash(bot, 40)
	wash := func(id uuid.UUID) int {
		view, _, err := room.Snapshot()
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range view.Seats {
			if s.ID == id.String() {
				return s.PlaymatWash
			}
		}
		t.Fatalf("no seat %s", id)
		return 0
	}
	if got := wash(human); got != 0 {
		t.Errorf("a wash with no playmat reached the wire: %d", got)
	}
	room.SetPlaymat(human, "/playmats/aaaa")
	if got := wash(human); got != 70 {
		t.Errorf("wash = %d, want 70", got)
	}
	if got := wash(bot); got != 0 {
		t.Errorf("a bot seat carries wash %d", got)
	}
	room.SetPlaymatWash(human, 0)
	if got := wash(human); got != 0 {
		t.Errorf("after clearing the wash: %d", got)
	}
}
