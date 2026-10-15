package ws

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// loop_replay_test.go — ADR 0193 (#3382, CR 732.2) at the room: a loop
// is played once by hand while it is recorded, proposed, answered by
// every other seat, and replayed through the engine as one commit.
//
// The fixture is Basalt Monolith ("{T}: Add {C}{C}{C}. {3}: Untap this
// artifact"), a deterministic untap loop with no net mana. Mesmeric Orb
// ("Whenever a permanent becomes untapped, that permanent's controller
// mills a card") makes every iteration move a hidden card, and Rings of
// Brighthearth ("you may pay {2}, copy that ability") puts a prompt and
// a copy — objects born inside the iteration — into it.

const (
	loopMonolithOracle = "6b8cf2a0-b045-4d91-9d91-c602d40c6237"
	loopRingsOracle    = "bbf9494c-c4bb-4d36-98fe-8387846b342e"
	loopOrbOracle      = "03efb4f3-b8e2-4441-824f-886dc40712c4"
)

type loopTable struct {
	room     *Room
	g        *game.Game
	seats    []uuid.UUID
	monolith uuid.UUID
	rings    bool
}

type loopTableOpts struct {
	seats int
	rings bool
	orb   bool
	float int
}

// newLoopTable is a room at the first seat's precombat main with a
// Basalt Monolith (and the asked-for extras) under the first seat.
func newLoopTable(t *testing.T, o loopTableOpts) *loopTable {
	t.Helper()
	if o.seats == 0 {
		o.seats = 2
	}
	g := game.NewGame()
	for i := range o.seats {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 40 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 7))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	lt := &loopTable{g: g, rings: o.rings}
	for _, p := range g.Seats {
		lt.seats = append(lt.seats, p.ID)
	}
	me := lt.seats[0]
	spawn := func(name, oracle string) uuid.UUID {
		ids, err := g.SpawnCards(me, me, game.ZoneBattlefield, game.Card{Name: name, TypeLine: "Artifact", OracleID: oracle}, 1)
		if err != nil {
			t.Fatalf("spawn %s: %v", name, err)
		}
		return ids[0]
	}
	lt.monolith = spawn("Basalt Monolith", loopMonolithOracle)
	if o.rings {
		spawn("Rings of Brighthearth", loopRingsOracle)
	}
	if o.orb {
		spawn("Mesmeric Orb", loopOrbOracle)
	}
	if o.float > 0 {
		g.WithWriteLock(func() {
			for range o.float {
				g.Seats[0].ManaPool.AddMana(game.ManaToken{Color: "C", Source: uuid.New()})
			}
		})
	}
	lt.room = NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), t.TempDir())
	return lt
}

// act dispatches an action as seat, through the room as the hub does.
func (lt *loopTable) act(seat uuid.UUID, typ actions.Type, params any) error {
	var raw json.RawMessage
	if params != nil {
		raw, _ = json.Marshal(params)
	}
	return lt.dispatch(actions.Action{Type: typ, Player: seat, Caller: seat, Params: raw})
}

func (lt *loopTable) dispatch(a actions.Action) error {
	fn := func() error { return actions.Dispatch(lt.room.Game, a) }
	tag := ActionTag{Type: a.Type, Actor: a.Caller}
	var err error
	if actions.MintsNoUndo(lt.room.Game, a.Type) {
		_, _, err = lt.room.ApplyExternalAction(tag, fn)
	} else {
		_, _, err = lt.room.ApplyAction(tag, fn)
	}
	return err
}

// move dispatches seat's first legal move pick accepts.
func (lt *loopTable) move(t *testing.T, seat uuid.UUID, pick func(legal.Move) bool) {
	t.Helper()
	moves := legal.EnumerateFor(lt.g, seat)
	for _, m := range moves {
		if pick(m) {
			if err := lt.dispatch(actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params}); err != nil {
				t.Fatalf("%s: %v", m.Label, err)
			}
			return
		}
	}
	var labels []string
	for _, m := range moves {
		labels = append(labels, m.Label)
	}
	t.Fatalf("no such move for %v among %q", seat, labels)
}

func (lt *loopTable) state(fn func()) { lt.g.ReadSnapshot(fn) }

func (lt *loopTable) holder() uuid.UUID {
	var id uuid.UUID
	lt.state(func() { id = lt.g.Seats[lt.g.Turn.PriorityHolder].ID })
	return id
}

func (lt *loopTable) stackSize() int {
	n := 0
	lt.state(func() { n = len(lt.g.StackMeta) })
	return n
}

func (lt *loopTable) monolithTapped() bool {
	tapped := false
	lt.state(func() {
		for _, c := range lt.g.Battlefield.Cards {
			if c.InstanceID == lt.monolith {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}

func (lt *loopTable) library(i int) int {
	n := 0
	lt.state(func() { n = lt.g.Seats[i].Library.Size() })
	return n
}

func (lt *loopTable) pool() int {
	n := 0
	lt.state(func() { n = lt.g.Seats[0].ManaPool.Total() })
	return n
}

func isPay(m legal.Move) bool {
	var p struct {
		Apply *bool `json:"apply"`
	}
	_ = json.Unmarshal(m.Params, &p)
	return p.Apply != nil && *p.Apply
}

// iteration plays one iteration of the Monolith loop by hand: tap it
// for {C}{C}{C}, pay {3} to untap it, and pass (and pay the Rings) until
// the stack is empty and the first seat has priority again.
func (lt *loopTable) iteration(t *testing.T) {
	t.Helper()
	me := lt.seats[0]
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindMana && m.Source == lt.monolith })
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindActivate && m.Source == lt.monolith })
	lt.settle(t)
}

// settle answers prompts (paying when it can) and passes until the
// stack is empty and the first seat holds priority.
func (lt *loopTable) settle(t *testing.T) {
	t.Helper()
	for range 60 {
		owed := false
		for _, s := range lt.seats {
			moves := legal.EnumerateFor(lt.g, s)
			if len(moves) == 0 || moves[0].Kind != legal.KindChoice {
				continue
			}
			owed = true
			lt.move(t, s, func(m legal.Move) bool { return m.Kind == legal.KindChoice && (!lt.rings || isPay(m)) })
			break
		}
		if owed {
			continue
		}
		if lt.stackSize() == 0 && lt.holder() == lt.seats[0] {
			return
		}
		lt.move(t, lt.holder(), func(m legal.Move) bool { return m.Kind == legal.KindPass })
	}
	t.Fatal("the iteration did not settle")
}

// record plays one iteration while recording it and stops.
func (lt *loopTable) record(t *testing.T) *game.LoopRecording {
	t.Helper()
	me := lt.seats[0]
	if err := lt.act(me, actions.TypeLoopRecordStart, nil); err != nil {
		t.Fatalf("loop_record_start: %v", err)
	}
	lt.iteration(t)
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err != nil {
		t.Fatalf("loop_record_stop: %v", err)
	}
	r := lt.g.CurrentLoopRecording()
	if r == nil || r.Status != game.LoopStatusReady {
		t.Fatalf("the recording did not close: %+v\n%s", r, lt.logText())
	}
	return r
}

// answerAll has every other seat accept the outstanding proposal.
func (lt *loopTable) answerAll(t *testing.T) {
	t.Helper()
	for range len(lt.seats) {
		c := lt.proposalPrompt()
		if c == nil {
			return
		}
		if err := lt.act(c.Chooser, actions.TypeResolveLoopProposal, map[string]any{"choice_id": c.ID.String(), "accept": true}); err != nil {
			t.Fatalf("accept: %v", err)
		}
	}
}

func (lt *loopTable) proposalPrompt() *game.PendingChoice {
	var out *game.PendingChoice
	lt.state(func() {
		for _, c := range lt.g.PendingChoices {
			if c.Kind == game.PendingChoiceLoopProposal {
				cp := *c
				out = &cp
			}
		}
	})
	return out
}

func (lt *loopTable) view() protocol.GameView {
	v, _, err := lt.room.Snapshot()
	if err != nil {
		panic(err)
	}
	return v
}

func (lt *loopTable) logText() string {
	var b strings.Builder
	for _, e := range lt.view().Log {
		b.WriteString(e.Text + "\n")
	}
	return b.String()
}

func (lt *loopTable) loopLines() []string {
	var out []string
	for _, e := range lt.view().Log {
		if e.Kind == protocol.LogLoop {
			out = append(out, e.Text)
		}
	}
	return out
}

func (lt *loopTable) undoDepth() int {
	lt.room.mu.Lock()
	defer lt.room.mu.Unlock()
	return len(lt.room.undoStack)
}

func lastLine(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// The whole flow: a recorded Monolith-and-Orb iteration replayed five
// times does what five iterations by hand do, as one undo entry and one
// log line, and the undo takes all five back.
func TestLoopReplayEqualsTheSameIterationsByHand(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{orb: true})
	byHand := newLoopTable(t, loopTableOpts{orb: true})

	r := lt.record(t)
	if len(r.Steps) != 6 {
		t.Errorf("recorded %d steps, want 6 (mana, untap, four passes): %+v", len(r.Steps), r.Steps)
	}
	if got := lt.g.LoopNetText(r, 1); !strings.Contains(got, "library -1") {
		t.Errorf("net change %q does not show the Orb's mill", got)
	}
	byHand.iteration(t)

	before := lt.undoDepth()
	activations := strings.Count(lt.logText(), "activated")
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 5}); err != nil {
		t.Fatalf("loop_propose: %v", err)
	}
	if lt.proposalPrompt() == nil || lt.proposalPrompt().Chooser != lt.seats[1] {
		t.Fatal("the other seat was not asked")
	}
	lt.answerAll(t)
	for range 5 {
		byHand.iteration(t)
	}

	if lt.g.CurrentLoopRecording() != nil {
		t.Fatal("the recording outlived its replay")
	}
	if got, want := lt.library(0), byHand.library(0); got != want {
		t.Errorf("library after the replay: %d, by hand %d", got, want)
	}
	if lt.pool() != byHand.pool() || lt.monolithTapped() || lt.stackSize() != 0 || lt.holder() != lt.seats[0] {
		t.Errorf("board after the replay: pool %d (by hand %d), tapped %v, stack %d, holder %v",
			lt.pool(), byHand.pool(), lt.monolithTapped(), lt.stackSize(), lt.holder())
	}
	if got := lt.undoDepth() - before; got != 1 {
		t.Errorf("the replay left %d undo entries, want 1", got)
	}
	if got := strings.Count(lt.logText(), "activated"); got != activations {
		t.Errorf("the replay wrote %d activation lines, want them folded into one line", got-activations)
	}
	line := lastLine(lt.loopLines())
	if !strings.Contains(line, "repeats Basalt Monolith ×5") || !strings.Contains(line, "library -5") {
		t.Errorf("replay line %q", line)
	}

	// One undo takes the whole replay back, from the proposer's budget.
	if _, _, err := lt.room.Undo(lt.seats[1]); err == nil {
		t.Error("the other seat undid the proposer's replay")
	}
	budget := 0
	lt.state(func() { budget = lt.g.Seats[0].UndosRemaining })
	if _, _, err := lt.room.Undo(lt.seats[0]); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	after := 0
	lt.state(func() { after = lt.g.Seats[0].UndosRemaining })
	if got, want := lt.library(0), byHand.library(0)+5; got != want {
		t.Errorf("library after undoing the replay: %d, want %d", got, want)
	}
	if after != budget-1 {
		t.Errorf("undoing the replay cost %d undos, want 1", budget-after)
	}
	if lt.g.CurrentLoopRecording() != nil {
		t.Error("undoing the replay brought the recording back")
	}
}

// CR 732.2b: each other seat is asked in turn order after the proposer,
// and one that shortens ends the replay before their own choice, with
// priority.
func TestLoopProposalAsksInTurnOrderAndShortens(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{seats: 3})
	r := lt.record(t)
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 10}); err != nil {
		t.Fatalf("loop_propose: %v", err)
	}
	first := lt.proposalPrompt()
	if first == nil || first.Chooser != lt.seats[1] {
		t.Fatalf("first asked %+v, want seat 2", first)
	}
	if err := lt.act(lt.seats[1], actions.TypeResolveLoopProposal, map[string]any{"choice_id": first.ID.String(), "accept": true}); err != nil {
		t.Fatalf("accept: %v", err)
	}
	second := lt.proposalPrompt()
	if second == nil || second.Chooser != lt.seats[2] {
		t.Fatalf("second asked %+v, want seat 3", second)
	}
	mine := r.LoopChoicePoints(lt.seats[2])
	if len(mine) == 0 {
		t.Fatal("seat 3 has no choice in the iteration")
	}
	// Not one of their own choices: refused.
	if err := lt.act(lt.seats[2], actions.TypeResolveLoopProposal, map[string]any{
		"choice_id": second.ID.String(), "shorten": map[string]int{"iteration": 3, "step": 0},
	}); err == nil {
		t.Fatal("seat 3 shortened the loop at the proposer's step")
	}
	library := lt.library(0)
	if err := lt.act(lt.seats[2], actions.TypeResolveLoopProposal, map[string]any{
		"choice_id": second.ID.String(), "shorten": map[string]int{"iteration": 3, "step": mine[0]},
	}); err != nil {
		t.Fatalf("shorten: %v", err)
	}
	if lt.g.CurrentLoopRecording() != nil {
		t.Fatal("the shortened proposal was not replayed")
	}
	if lt.holder() != lt.seats[2] {
		t.Errorf("priority is with %v, want the seat that shortened", lt.holder())
	}
	if lt.stackSize() != 1 || !lt.monolithTapped() {
		t.Errorf("stack %d, Monolith tapped %v at the shortened end; want the untap ability waiting", lt.stackSize(), lt.monolithTapped())
	}
	if lt.library(0) != library {
		t.Error("the Monolith loop without an Orb moved a card")
	}
	line := lastLine(lt.loopLines())
	if !strings.Contains(line, "×2") || !strings.Contains(line, "shortened by P3 to iteration 3") {
		t.Errorf("replay line %q", line)
	}
}

// ADR 0193 §5: a replay that diverges rolls back to the end of the last
// complete iteration. With Rings of Brighthearth paid every time, each
// iteration costs {2}; nine floating mana pay for the demonstration and
// three more, and the fourth cannot pay.
func TestLoopReplayDivergenceRollsBackToTheLastFullIteration(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{rings: true, float: 9})
	r := lt.record(t)
	if lt.pool() != 7 {
		t.Fatalf("pool %d after the demonstration, want 7", lt.pool())
	}
	var born bool
	for _, s := range r.Steps {
		born = born || strings.Contains(s.Params, game.LoopRefPrefix)
	}
	if !born {
		t.Errorf("no step names an object born in the iteration: %+v", r.Steps)
	}
	before := lt.undoDepth()
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 10}); err != nil {
		t.Fatalf("loop_propose: %v", err)
	}
	lt.answerAll(t)
	if lt.pool() != 1 {
		t.Errorf("pool %d after the replay, want 1 (three iterations at {2} each)", lt.pool())
	}
	if lt.monolithTapped() || lt.stackSize() != 0 || lt.holder() != lt.seats[0] {
		t.Error("the table was left mid-iteration")
	}
	if got := lt.undoDepth() - before; got != 1 {
		t.Errorf("%d undo entries, want 1", got)
	}
	line := lastLine(lt.loopLines())
	if !strings.Contains(line, "×3") || !strings.Contains(line, "stopped after 3 of 10") {
		t.Errorf("replay line %q", line)
	}
}

// A recording that leaves Basalt Monolith tapped is not a loop: it is
// discarded, and the log says why.
func TestLoopRecordingThatDoesNotCloseIsDiscarded(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	me := lt.seats[0]
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err == nil {
		t.Error("stopped a recording that was never started")
	}
	if err := lt.act(me, actions.TypeLoopRecordStart, nil); err != nil {
		t.Fatal(err)
	}
	if err := lt.act(lt.seats[1], actions.TypeLoopRecordStart, nil); err == nil {
		t.Error("a second recording started")
	}
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err == nil {
		t.Error("stopped an empty recording")
	}
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindMana && m.Source == lt.monolith })
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err != nil {
		t.Fatal(err)
	}
	if lt.g.CurrentLoopRecording() != nil {
		t.Fatal("a recording that does not close was kept")
	}
	if line := lastLine(lt.loopLines()); !strings.Contains(line, "Basalt Monolith is tapped at the end and was untapped at the start") {
		t.Errorf("discard line %q", line)
	}
}

// Owner decision 2: a proposal is 1 to 1000 iterations, and a replay
// stops at its step budget, keeping the iterations it completed.
func TestLoopProposalCapsAndReplayBudget(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	lt.record(t)
	for _, n := range []int{0, -1, game.MaxLoopShortcutIterations + 1} {
		if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": n}); err == nil {
			t.Errorf("proposed %d iterations", n)
		}
	}
	if err := lt.act(lt.seats[1], actions.TypeLoopPropose, map[string]int{"count": 3}); err == nil {
		t.Error("a seat proposed someone else's loop")
	}
	lt.room.loopBudget = 9 // four steps an iteration: two iterations fit
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 5}); err != nil {
		t.Fatal(err)
	}
	lt.answerAll(t)
	line := lastLine(lt.loopLines())
	if !strings.Contains(line, "×2") || !strings.Contains(line, "stopped after 2 of 5: the replay reached its limit of steps") {
		t.Errorf("replay line %q", line)
	}
	if lt.monolithTapped() || lt.stackSize() != 0 {
		t.Error("the budget left the table mid-iteration")
	}
}

// Owner decision 4: hidden information. A step that used a card from a
// hand (it was cast, and is in the graveyard now) cannot be taken again:
// the replay stops in its first iteration and commits nothing.
func TestLoopReplayOfAHiddenCardStopsAndCommitsNothing(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	me := lt.seats[0]
	spark, err := lt.g.SpawnCards(me, me, game.ZoneHand, game.Card{Name: "Test Spark", TypeLine: "Instant", ManaCost: "{0}"}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := lt.act(me, actions.TypeLoopRecordStart, nil); err != nil {
		t.Fatal(err)
	}
	if err := lt.act(me, actions.TypeCastSpell, map[string]any{"instance_id": spark[0].String(), "from_zone": "hand"}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	lt.settle(t)
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err != nil {
		t.Fatal(err)
	}
	before := lt.undoDepth()
	if err := lt.act(me, actions.TypeLoopPropose, map[string]int{"count": 3}); err != nil {
		t.Fatal(err)
	}
	lt.answerAll(t)
	if lt.undoDepth() != before {
		t.Error("a replay that completed nothing left an undo entry")
	}
	line := lastLine(lt.loopLines())
	if !strings.Contains(line, "×0") || !strings.Contains(line, "stopped after 0 of 3") || !strings.Contains(line, "was refused") {
		t.Errorf("replay line %q", line)
	}
	if lt.g.CurrentLoopRecording() != nil {
		t.Error("the recording outlived its replay")
	}
}

// Owner decision 4: randomness. A shuffle inside the loop is drawn
// fresh every iteration, and no choice depends on it, so the replay
// runs to the end.
func TestLoopReplayWithAShuffleRuns(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	me := lt.seats[0]
	if err := lt.act(me, actions.TypeLoopRecordStart, nil); err != nil {
		t.Fatal(err)
	}
	if err := lt.act(me, actions.TypeShuffleLibrary, nil); err != nil {
		t.Fatalf("shuffle: %v", err)
	}
	if err := lt.act(me, actions.TypeLoopRecordStop, nil); err != nil {
		t.Fatal(err)
	}
	if err := lt.act(me, actions.TypeLoopPropose, map[string]int{"count": 4}); err != nil {
		t.Fatal(err)
	}
	lt.answerAll(t)
	if line := lastLine(lt.loopLines()); !strings.Contains(line, "×4") || strings.Contains(line, "stopped") {
		t.Errorf("replay line %q", line)
	}
}

// ADR 0193 §5: the CR 732 breaker sits out a replay. With the threshold
// at one, a single untap by hand trips it; five replayed do not.
func TestLoopReplayDoesNotTripTheLoopBreaker(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	lt.record(t)
	lt.g.WithWriteLock(func() { lt.g.LoopThreshold = 1 })
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 5}); err != nil {
		t.Fatal(err)
	}
	lt.answerAll(t)
	if line := lastLine(lt.loopLines()); !strings.Contains(line, "×5") {
		t.Fatalf("replay line %q", line)
	}
	if lt.g.CurrentLoopNotice() != nil {
		t.Error("the breaker raised its notice during the replay")
	}
	lt.state(func() {
		for _, c := range lt.g.PendingChoices {
			if c.Kind == game.PendingChoiceLoopShortcut {
				t.Error("the breaker queued its shortcut prompt during the replay")
			}
		}
	})
	// The control: the same threshold trips on one iteration by hand.
	me := lt.seats[0]
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindMana && m.Source == lt.monolith })
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindActivate && m.Source == lt.monolith })
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindPass })
	lt.move(t, lt.seats[1], func(m legal.Move) bool { return m.Kind == legal.KindPass })
	if lt.g.CurrentLoopNotice() == nil {
		t.Error("the control iteration did not trip the breaker, so the test proves nothing")
	}
}

// ADR 0193 §2: an undo while recording ends the recording; cancel
// withdraws a proposal and its prompt.
func TestLoopRecordingEndsOnUndoAndCancel(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	me := lt.seats[0]
	if err := lt.act(me, actions.TypeLoopRecordStart, nil); err != nil {
		t.Fatal(err)
	}
	lt.move(t, me, func(m legal.Move) bool { return m.Kind == legal.KindMana && m.Source == lt.monolith })
	if _, _, err := lt.room.Undo(me); err != nil {
		t.Fatal(err)
	}
	if lt.g.CurrentLoopRecording() != nil {
		t.Error("the recording survived an undo")
	}

	lt.record(t)
	if err := lt.act(me, actions.TypeLoopPropose, map[string]int{"count": 3}); err != nil {
		t.Fatal(err)
	}
	if err := lt.act(lt.seats[1], actions.TypeLoopRecordCancel, nil); err == nil {
		t.Error("another seat cancelled the recording")
	}
	if err := lt.act(me, actions.TypeLoopRecordCancel, nil); err != nil {
		t.Fatal(err)
	}
	if lt.g.CurrentLoopRecording() != nil || lt.proposalPrompt() != nil {
		t.Error("cancel left the recording or its prompt behind")
	}
}

// The wire: every seat sees the recording, and the prompt carries the
// proposal and the chooser's own choice points.
func TestLoopRecordingIsOnTheWire(t *testing.T) {
	lt := newLoopTable(t, loopTableOpts{})
	lt.record(t)
	v := protocol.FilterViewFor(lt.view(), lt.seats[1].String())
	if v.LoopRecording == nil || v.LoopRecording.Status != "ready" || v.LoopRecording.Steps != 4 || v.LoopRecording.Recorder != lt.seats[0].String() {
		t.Fatalf("loop_recording %+v", v.LoopRecording)
	}
	if err := lt.act(lt.seats[0], actions.TypeLoopPropose, map[string]int{"count": 7}); err != nil {
		t.Fatal(err)
	}
	v = protocol.FilterViewFor(lt.view(), lt.seats[1].String())
	var lp *protocol.LoopProposalView
	for _, c := range v.PendingChoices {
		if c.Kind == string(game.PendingChoiceLoopProposal) {
			lp = c.LoopProposal
		}
	}
	if lp == nil || lp.Count != 7 || len(lp.Steps) != 4 || len(lp.ChoicePoints) == 0 || lp.End.Iteration != 7 {
		t.Fatalf("loop_proposal %+v", lp)
	}
	for _, i := range lp.ChoicePoints {
		if lp.Steps[i].Actor != lt.seats[1].String() {
			t.Errorf("choice point %d is not the chooser's: %+v", i, lp.Steps[i])
		}
	}
}
