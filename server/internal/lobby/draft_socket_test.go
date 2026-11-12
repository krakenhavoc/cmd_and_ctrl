package lobby

import (
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The measure of #5698: one person and seven bots, a whole pick round
// through a real draft socket, read slowly. The bots pick inside the
// person's pick (draft.advance), so each pick changes the draft once
// and sends the person one view. A person picks from the view they were
// sent, so at most one view is ever waiting for them: their own picks
// cannot fill the send buffer, however slowly they read.
func TestAPickRoundOverASlowDraftSocketIsOneViewAPick(t *testing.T) {
	s := newDraftStack(t)
	_, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
	id := out.Draft.ID
	conn := s.dial(t, id, out.Token)
	conn.next()
	s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", out.Token, "")
	cur := conn.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	if len(cur.Players) != 8 {
		t.Fatalf("%d seats, want 8", len(cur.Players))
	}
	picks, most := 0, 0
	for cur.Round == 1 && cur.State == draft.StatePicking {
		conn.pick(cur)
		picks++
		views := 0
		for {
			// The slow reader.
			time.Sleep(5 * time.Millisecond)
			v := conn.next()
			views++
			if len(v.Picks) == picks {
				cur = v
				break
			}
		}
		most = max(most, views)
	}
	if picks != 14 {
		t.Fatalf("%d picks in the first round, want 14", picks)
	}
	if most != 1 {
		t.Fatalf("a pick was answered by %d views, want one each", most)
	}
	if st := s.lobby.DraftRoom(id).Status(); st.Connected != 1 {
		t.Fatalf("after the round %d seats are connected, want the person's", st.Connected)
	}
	t.Logf("%d picks, at most %d view a pick", picks, most)
}

// The burst #5696 met in a test: every pick of a whole draft made
// server-side in a tight loop while the person's socket is not read.
// The views are the draft's state, so the hub keeps only the newest
// one not yet written instead of queueing all of them: the socket
// stays up, and reading it reaches the end of the picks.
func TestABurstOfDraftViewsDoesNotDropTheDrafter(t *testing.T) {
	s := newDraftStack(t)
	_, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn"}`)
	id := out.Draft.ID
	seat := s.draftSeatOf(t, out.Token)
	conn := s.dial(t, id, out.Token)
	conn.next()
	s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", out.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, seat)
	want, _ := room.View(seat)
	if want.State != draft.StateDone {
		t.Fatalf("after every pick the draft is %q", want.State)
	}
	got := conn.until("the end of the picks", func(v draft.View) bool { return v.Seq >= want.Seq })
	if got.State != draft.StateDone || len(got.Pool) != 42 {
		t.Fatalf("the last view: %q with a pool of %d", got.State, len(got.Pool))
	}
	if st := room.Status(); st.Connected != 1 {
		t.Fatalf("after the burst %d seats are connected, want the person's", st.Connected)
	}
	views := 0
	for _, raw := range conn.raw {
		var f protocol.Frame
		if json.Unmarshal(raw, &f) == nil && f.Kind == protocol.KindDraft {
			views++
		}
	}
	t.Logf("42 picks made server-side; the socket was written %d views", views)
}

// A brief disconnect is not walking away (#5698). A match that stalls
// just after the person's socket dropped is recorded as a draw, as it
// is for a person on the page; the draft is not closed. The grace is
// the pick timer's: a seat is away once it has had no socket for
// draft.AwayGrace.
func TestABriefDisconnectDoesNotAbandonADraft(t *testing.T) {
	s := newDraftStack(t)
	s.withRealBots(t)
	room, seat := s.draftForTheRounds(t, 0)
	draftEventually(t, "the draft finishes on draws", func() bool {
		_, phase := s.roundOf(room, seat)
		return phase == draft.PhaseFinished
	})
	v, _ := room.View(seat)
	if st := v.Play.Standings[v.You]; st.Draws != draft.MatchRounds {
		t.Fatalf("the record = %+v, want three draws", st)
	}
	if s.lobby.DraftRoom(room.ID()) == nil {
		t.Fatal("the draft was closed")
	}
}

// Past the grace, the person has walked away: a stalled match closes
// the draft, as it does for a person who never opened the page.
func TestADisconnectPastTheGraceAbandonsADraft(t *testing.T) {
	s := newDraftStack(t)
	s.withRealBots(t)
	room, _ := s.draftForTheRounds(t, draft.AwayGrace)
	draftEventually(t, "the abandoned draft closes", func() bool { return s.lobby.DraftRoom(room.ID()) == nil })
}

// frozenClock is a draft room's clock, moved by hand.
type frozenClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *frozenClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *frozenClock) Add(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// draftForTheRounds is a two-seat draft whose person was on its page,
// picked everything and then lost their socket; the room's clock is
// then moved on by away. Their deck goes in last, so the rounds' tables
// open with nobody connected, and stall after 150 ms.
func (s *draftStack) draftForTheRounds(t *testing.T, away time.Duration) (*draft.Room, uuid.UUID) {
	t.Helper()
	_, out := s.call(t, http.MethodPost, "/draft", "", `{"set":"syn","seats":2}`)
	id := out.Draft.ID
	seat := s.draftSeatOf(t, out.Token)
	s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", out.Token, "")
	room := s.lobby.DraftRoom(id)
	pickEverything(t, room, seat)
	clock := &frozenClock{now: time.Now()}
	room.SetClock(clock.Now, nil)
	conn := s.dial(t, id, out.Token)
	conn.next()
	_ = conn.conn.Close()
	draftEventually(t, "the socket is gone", func() bool { return room.Status().Connected == 0 })
	clock.Add(away)
	s.lobby.SetDraftMatchLimits(DraftMatchLimits{Idle: 150 * time.Millisecond})
	if err := room.SubmitDeck(seat, deckOf(t, room, seat)); err != nil {
		t.Fatal(err)
	}
	return room, seat
}
