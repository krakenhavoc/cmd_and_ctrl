package aiseat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// search_hatch_test.go covers the second half of #544: the runner's
// escape from a prompt it cannot answer.
//
// Enumerating only picks the engine accepts (see
// legal/search_validate_test.go) fixes the Myriad Landscape wedge at
// its source. It does not fix the SHAPE of the trap, which is that a
// seat owing a pending choice is enumerated that choice's answers and
// nothing else — so PassIndex is -1 for as long as the prompt is
// open, and Runner.step's forced-pass rescue had nothing to force and
// went to sleep still holding the table.
//
// The hatch is legal.Move.AlwaysLegal: a choice kind declares the one
// answer nothing can refuse, and a runner out of rejections takes it
// rather than returning. A bot that fails to find is playing badly. A
// bot that stops is holding three other people hostage.

// millOnPick answers a search prompt with a real "take" answer and,
// on its way out, moves the card it just picked out of the library —
// so the answer it returns is stale by the time the runner dispatches
// it, and the engine refuses it with ErrCardNotFound.
//
// That is not a contrivance; it is the ordinary race the enumerator
// cannot close (another seat's trigger mills your library between the
// frame and the click), and it is the reason AlwaysLegal claims more
// than "legal at enumeration time". It never returns the always-legal
// answer voluntarily, so if the prompt ever clears, the runner is
// what cleared it.
type millOnPick struct {
	t    *testing.T
	room *ws.Room
	seat uuid.UUID

	mu     sync.Mutex
	takes  int
	safes  int
	milled int
}

func (p *millOnPick) Name() string { return "mill-on-pick" }

func (p *millOnPick) Decide(_ context.Context, in aiseat.Input) (aiseat.Decision, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, m := range in.Moves {
		if m.AlwaysLegal || !strings.Contains(m.Label, ": take ") {
			continue
		}
		p.takes++
		for _, id := range moveCardIDs(p.t, m) {
			if _, _, err := p.room.ApplyExternal(func() error {
				return p.room.Game.MoveCardByID(
					game.ZoneRef{Kind: game.ZoneLibrary, Owner: p.seat},
					game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.seat},
					id)
			}); err != nil {
				p.t.Errorf("mill %s: %v", id, err)
			}
			p.milled++
		}
		return aiseat.Decision{Index: i, Reason: "take (and break it on the way out)"}, nil
	}
	// No take on offer. If that is because only the always-legal
	// answer is left, decline rather than take it: the point of the
	// test is that the RUNNER reaches for it.
	for _, m := range in.Moves {
		if m.AlwaysLegal && m.Kind == legal.KindChoice {
			p.safes++
		}
	}
	return aiseat.Decision{Index: aiseat.Decline, Reason: "nothing this policy wants"}, nil
}

func moveCardIDs(t *testing.T, m legal.Move) []uuid.UUID {
	t.Helper()
	var p struct {
		CardIDs []string `json:"card_ids"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params of %q: %v", m.Label, err)
	}
	out := make([]uuid.UUID, 0, len(p.CardIDs))
	for _, s := range p.CardIDs {
		id, err := uuid.Parse(s)
		if err != nil {
			t.Fatalf("card id %q: %v", s, err)
		}
		out = append(out, id)
	}
	return out
}

// newMyriadRoom builds a two-seat room on myriadDeck with the active
// seat's Myriad Landscape already cracked and its search prompt open.
// The runner is not started yet, so the board can be set up without
// the room lock.
func newMyriadRoom(t *testing.T, forests, islands int) (*ws.Room, *game.Player) {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), myriadDeck(uuid.Nil, forests, islands)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 12))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[0]
	lib := game.ZoneRef{Kind: game.ZoneLibrary, Owner: me.ID}
	bf := game.ZoneRef{Kind: game.ZoneBattlefield}

	landscape := findInLibrary(g, me, "Myriad Landscape")
	if landscape == uuid.Nil {
		t.Fatal("no Myriad Landscape in the library")
	}
	if err := g.MoveCardByID(lib, bf, landscape); err != nil {
		t.Fatalf("move landscape: %v", err)
	}
	if err := g.TapCard(landscape, false); err != nil {
		t.Fatalf("untap landscape: %v", err)
	}
	for i := 0; i < 2; i++ {
		if err := g.MoveCardByID(lib, bf, findInLibrary(g, me, "Forest")); err != nil {
			t.Fatalf("move forest: %v", err)
		}
	}

	moves := legal.EnumerateFor(g, me.ID)
	act := -1
	for i, m := range moves {
		if m.Type == legal.TypeActivateAbility && m.Source == landscape {
			act = i
			break
		}
	}
	if act < 0 {
		t.Fatalf("the enumerator never offered Myriad Landscape's ability (%d moves)", len(moves))
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(moves[act].Type), Player: me.ID, Caller: me.ID, Params: moves[act].Params,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 20 && searchChoice(g, me.ID) == nil; i++ {
		holder := g.Seats[g.Turn.PriorityHolder].ID
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.TypePassPriority, Player: holder, Caller: holder,
		}); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if searchChoice(g, me.ID) == nil {
		t.Fatal("the search prompt never opened")
	}
	return ws.NewRoom(g, testLogger(), ""), me
}

// TestRunnerTakesTheAlwaysLegalAnswerWhenEveryPickIsRejected: a seat
// whose every considered answer bounces still advances, because the
// runner reaches for the choice's always-legal answer once its
// rejection budget is spent. Before #544 this returned and slept,
// and the prompt — and the table — stayed open forever.
func TestRunnerTakesTheAlwaysLegalAnswerWhenEveryPickIsRejected(t *testing.T) {
	// simic-ramp's shape — 7 Forest, 5 Island in 99 cards — which is
	// what the field replay wedged on, on turn 2 with a full library.
	room, me := newMyriadRoom(t, 7, 5)
	g := room.Game
	if pi := aiseat.PassIndex(legal.EnumerateFor(g, me.ID)); pi >= 0 {
		t.Fatal("scenario is wrong: a seat owing a choice must not be offered a pass")
	}

	pol := &millOnPick{t: t, room: room, seat: me.ID}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := aiseat.Start(ctx, room, me.ID, pol, aiseat.Config{
		MaxThink:              time.Second,
		MaxConsecutiveRejects: 2,
	}, nil, testLogger())

	waitFor(t, "the search prompt to clear", 5*time.Second, func() bool {
		// Under the read lock: the runner is mutating this game from
		// its own goroutine.
		open := true
		g.ReadSnapshot(func() { open = searchChoice(g, me.ID) != nil })
		return !open
	})
	cancel()
	<-r.Done()

	st := r.Stats()
	if st.Rejected < 2 {
		t.Errorf("the scenario did not actually reject anything: %+v", st)
	}
	pol.mu.Lock()
	defer pol.mu.Unlock()
	if pol.takes < 2 {
		t.Errorf("the policy never insisted on a doomed pick (%d takes)", pol.takes)
	}
	if pol.safes > 0 {
		t.Errorf("the policy was offered only the always-legal answer %d times; "+
			"the prompt may have cleared for the wrong reason", pol.safes)
	}
}

// TestMyriadLandscapeFetchesTwoLands is the play-quality assertion,
// end to end and at a realistic library size.
//
// MaxExpansionPerSource is 12 and the old enumerator walked pick
// sizes in order, so a library holding more than twelve matching
// basics — every Commander mana base — offered single cards only and
// the bot took one land off a card that prints two. The budget is now
// spread across pick sizes, so the pairs are on offer and the
// heuristic (which sums cardValue over the pick) takes one.
func TestMyriadLandscapeFetchesTwoLands(t *testing.T) {
	// The other side of the cap: 40 matching basics, far more than
	// MaxExpansionPerSource, which is the shape that used to offer no
	// pair at all. This never wedged — it just fetched one land.
	room, me := newMyriadRoom(t, 20, 20)
	g := room.Game
	ch := searchChoice(g, me.ID)
	if len(ch.SearchCards) <= 12 {
		t.Fatalf("scenario is too small to test the budget: %d candidates", len(ch.SearchCards))
	}
	before := basicsOnBattlefield(g, me.ID)

	moves := legal.EnumerateFor(g, me.ID)
	pol := heuristic.New()
	d, err := pol.Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, me.ID.String()),
		Seat:  me.ID,
		Moves: moves,
	})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	chosen := moves[d.Index]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(chosen.Type), Player: me.ID, Caller: me.ID, Params: chosen.Params,
	}); err != nil {
		t.Fatalf("the heuristic's answer %q was refused: %v", chosen.Label, err)
	}
	if got := basicsOnBattlefield(g, me.ID) - before; got != 2 {
		t.Errorf("Myriad Landscape fetched %d land(s), want 2 — chose %q", got, chosen.Label)
	}
}

func basicsOnBattlefield(g *game.Game, seat uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == seat && strings.Contains(c.TypeLine, "Basic Land") {
			n++
		}
	}
	return n
}
