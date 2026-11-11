package lobby

import (
	"net/http"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/draft"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// POST /draft takes build_untimed beside build_seconds (#5700). The
// two are exclusive, and an absent pair is the 15-minute timer as ever.
func TestCreateDraftBuildTimerChoice(t *testing.T) {
	s := newDraftStack(t)
	for body, want := range map[string]int{
		`{"set":"syn","build_untimed":true}`:                       http.StatusCreated,
		`{"set":"syn","build_untimed":false}`:                      http.StatusCreated,
		`{"set":"syn","build_seconds":600}`:                        http.StatusCreated,
		`{"set":"syn","build_untimed":true,"build_seconds":600}`:   http.StatusBadRequest,
		`{"set":"syn","build_seconds":30}`:                         http.StatusBadRequest,
		`{"set":"syn","build_untimed":true,"build_seconds":99999}`: http.StatusBadRequest,
	} {
		if resp, out := s.call(t, http.MethodPost, "/draft", "", body); resp.StatusCode != want {
			t.Errorf("POST /draft %s: %d (%s), want %d", body, resp.StatusCode, out.Error, want)
		}
	}
	// The draft carries the choice.
	untimed := s.create(t, "", `{"set":"syn","build_untimed":true}`)
	if room := s.lobby.DraftRoom(untimed.Draft.ID); room == nil {
		t.Fatal("no room")
	} else {
		room.With(func(d *draft.Draft) {
			if !d.BuildUntimed() {
				t.Error("the draft is timed")
			}
		})
	}
	timed := s.create(t, "", `{"set":"syn"}`)
	s.lobby.DraftRoom(timed.Draft.ID).With(func(d *draft.Draft) {
		if d.BuildUntimed() || d.Config().BuildTimeout != 0 {
			t.Errorf("the default draft is untimed or has a length: %+v", d.Config())
		}
	})
}

// An untimed draft over the socket: no deadline in the build, and the
// rounds begin on the person's submit.
func TestAnUntimedDraftBuildsOverTheSocket(t *testing.T) {
	s := newDraftStack(t)
	out := s.create(t, "", `{"set":"syn","seats":2,"build_untimed":true,"pick_seconds":300}`)
	id, tok := out.Draft.ID, out.Token
	conn := s.dial(t, id, tok)
	conn.next()
	if resp, _ := s.call(t, http.MethodPost, "/draft/"+id.String()+"/start", tok, ""); resp.StatusCode != http.StatusOK {
		t.Fatalf("start: %d", resp.StatusCode)
	}
	cur := conn.until("a pack", func(v draft.View) bool { return len(v.Pack) > 0 })
	for picks := 0; cur.State == draft.StatePicking; {
		conn.pick(cur)
		picks++
		want := picks
		cur = conn.until("the next pack or the build", func(v draft.View) bool {
			return v.State == draft.StateDone || (len(v.Picks) == want && len(v.Pack) > 0)
		})
	}
	if cur.Phase != draft.PhaseBuilding || cur.Build == nil || !cur.Build.Untimed || cur.Build.Deadline != nil || cur.Build.Seconds != 0 {
		t.Fatalf("after the last pick: phase %q, build %+v", cur.Phase, cur.Build)
	}
	deck := protocol.DraftDeckPayload{Basics: map[string]int{"Forest": 17}}
	for _, c := range cur.Pool[:23] {
		deck.Main = append(deck.Main, c.ID)
	}
	conn.send(protocol.KindDraftDeck, deck)
	conn.until("the rounds", func(v draft.View) bool { return v.Phase == draft.PhaseRounds || v.Phase == draft.PhaseFinished })
}
