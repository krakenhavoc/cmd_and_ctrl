package game

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// loop_recording.go — demonstrate a loop once, replay it N times
// (ADR 0193, #3382, CR 732.2).
//
// CR 732.2a lets the player with priority propose a shortcut: "a
// sequence of game choices, for all players", repeated a number of
// times. This file is the game half of that conversation:
//
//   - Record. Any seat in the game asks to record. The recording itself
//     begins at the recorder's next quiet priority (holding priority, no
//     prompt open, no trigger waiting), at once if they are there now;
//     until then it is armed and captures nothing (#4736). From the
//     start on, every action any seat dispatches is a step
//     (actions.Dispatch writes them through BeginLoopStep /
//     CommitLoopStep). The game stores a step's params as JSON text and
//     never decodes them.
//   - Stop. At the recorder's next quiet priority, at once if they are
//     there now, the closure check (loopSignatureLocked,
//     compareLoopSignatures) says whether the board came back to where
//     it started. A recording that does not close is discarded, with
//     the first difference in the log.
//   - Propose and answer. The recorder names N; each other seat, in turn
//     order after them, accepts or shortens (CR 732.2b) through a
//     blocking loop_proposal prompt.
//   - Replay. Once the last seat has answered the recording is due, and
//     the room (ws.Room.ReplayLoop) replays it through actions.Dispatch
//     as one commit. The breaker of ADR 0055 sits out while it does
//     (loopReplaying).
//
// Everything here is plain data on Game.LoopRecording, so an undo clone
// and a snapshot carry it.

// LoopRecordingStatus is where a recording is in ADR 0193 §1's flow.
type LoopRecordingStatus string

const (
	// LoopStatusArmed: the recorder asked to record while they could not
	// start (a prompt was open partway through a resolution, or another
	// seat held priority). Nothing is captured yet; the recording begins
	// at their next quiet priority (TendLoopRecording, #4736).
	LoopStatusArmed LoopRecordingStatus = "armed"
	// LoopStatusRecording: the iteration is being played by hand.
	LoopStatusRecording LoopRecordingStatus = "recording"
	// LoopStatusReady: the recording closed; the recorder may propose.
	LoopStatusReady LoopRecordingStatus = "ready"
	// LoopStatusProposed: a count was proposed and the other seats are
	// being asked, one at a time.
	LoopStatusProposed LoopRecordingStatus = "proposed"
	// LoopStatusAccepted: every seat has answered. The room replays the
	// recording at the end of the commit that set this, so no frame
	// ever shows it.
	LoopStatusAccepted LoopRecordingStatus = "accepted"
)

// PendingChoiceLoopProposal is CR 732.2b's question to each other
// player: accept the proposed shortcut, or shorten it to a place where
// you will choose differently (ADR 0193 §4). Answered with the
// resolve_loop_proposal action. Blocking: the answer decides what the
// next N iterations are.
const PendingChoiceLoopProposal PendingChoiceKind = "loop_proposal"

// MaxLoopReplayDispatches is the step budget of one replay (owner
// decision 2): how many recorded steps a replay may dispatch, across
// every iteration, before it stops.
const MaxLoopReplayDispatches = 20000

// LoopRefPrefix starts a recorded reference to an object created inside
// the iteration ("@loop:permanent:2" is the third object born in it, a
// permanent). actions rewrites such IDs in a step's params when it is
// recorded and back when it is replayed.
const LoopRefPrefix = "@loop:"

// Loop object kinds, the second field of a reference.
const (
	LoopObjectStack     = "stack"
	LoopObjectPermanent = "permanent"
	LoopObjectTrigger   = "trigger"
	LoopObjectChoice    = "choice"
)

// Errors the loop actions return. Each reads as a sentence once the
// "game: " prefix is trimmed for the wire.
var (
	ErrLoopNotRecorder      = errors.New("game: only the player recording the loop may do that")
	ErrLoopNoRecording      = errors.New("game: there is no loop recording")
	ErrLoopAlreadyRecording = errors.New("game: a loop is already being recorded")
	ErrLoopNeedsPriority    = errors.New("game: you need priority to propose a loop")
	ErrLoopPromptOpen       = errors.New("game: answer the open prompt first")
	ErrLoopNotRecording     = errors.New("game: the loop is not being recorded")
	ErrLoopEmpty            = errors.New("game: nothing has been recorded yet")
	ErrLoopNotReady         = errors.New("game: the loop recording is not ready to propose")
	ErrLoopCount            = errors.New("game: the number of iterations is out of range")
	ErrLoopBoardChanged     = errors.New("game: the board has changed since the loop was recorded")
	ErrLoopShortenPlace     = errors.New("game: you can only shorten the loop at one of your own choices in it")
	ErrLoopShortenTooLate   = errors.New("game: the loop already ends before that point")
)

// LoopRecording is one demonstrated loop: who recorded it, where it is
// in the flow, the board it started from and every step of the
// iteration. Pure data, carried by Clone, RestoreFrom and the snapshot.
type LoopRecording struct {
	Recorder uuid.UUID           `json:"recorder"`
	Status   LoopRecordingStatus `json:"status"`
	// StopAtPriority is set when the recorder stopped while they could
	// not end the iteration (no priority, or a prompt open): the closure
	// check runs at their next quiet priority, the first place CR 732.2a
	// lets the sequence end (#4736). Recording goes on until then.
	StopAtPriority bool `json:"stopAtPriority,omitempty"`
	// StartSeq is the event seq when the recording began. It names the
	// recording, so a step drafted against one never lands on another.
	StartSeq uint64 `json:"startSeq,omitempty"`
	// Start is the loop signature at the start (ADR 0193 §3): what every
	// iteration must give back. Zero while the recording is armed.
	Start LoopSignature `json:"start"`
	// Totals is the allowed-to-differ part of the board at the start,
	// the baseline of the net change.
	Totals LoopTotals `json:"totals"`
	// StartObjects is every object that existed when recording began;
	// anything else a step names was created inside the iteration.
	StartObjects []uuid.UUID `json:"startObjects,omitempty"`
	// Births is the objects created inside the iteration, in the order
	// the recording saw them.
	Births []LoopObject `json:"births,omitempty"`
	// Steps is the iteration, every seat's choices in order.
	Steps []LoopStep `json:"steps,omitempty"`
	// Net is the change one iteration makes, set when it closes.
	Net []LoopDelta `json:"net,omitempty"`
	// Proposal is the outstanding CR 732.2a proposal, once made.
	Proposal *LoopProposal `json:"proposal,omitempty"`
}

// LoopStep is one recorded action. Params is the action's JSON params
// with every ID of an object born in the iteration rewritten to a
// LoopRefPrefix reference.
type LoopStep struct {
	// Actor is the seat that sent the action (the action's caller);
	// uuid.Nil for the admin.
	Actor  uuid.UUID `json:"actor,omitempty"`
	Admin  bool      `json:"admin,omitempty"`
	Player uuid.UUID `json:"player,omitempty"`
	Type   string    `json:"type"`
	Params string    `json:"params,omitempty"`
	// Label is the step as the table reads it: "Luke: activate Basalt
	// Monolith".
	Label string `json:"label,omitempty"`
	// Source is the name of the card the step is about, if any; the
	// replay's log line names the loop from it.
	Source string `json:"source,omitempty"`
	// Pending is the prompts open just before the step ("kind@seat"),
	// which a replay must find open at the same point.
	Pending []string `json:"pending,omitempty"`
}

// LoopObject is one object a step can name: a stack item, a permanent,
// a waiting trigger or a pending prompt.
type LoopObject struct {
	ID   uuid.UUID `json:"id"`
	Kind string    `json:"kind"`
}

// LoopPoint is a place in a proposed sequence: before step Step of
// iteration Iteration (both counted as ADR 0193 §4 does, iterations
// from 1, steps from 0). The whole proposal ends at {Count, len(Steps)}.
type LoopPoint struct {
	Iteration int `json:"iteration"`
	Step      int `json:"step"`
}

// before reports whether p is strictly earlier than q.
func (p LoopPoint) before(q LoopPoint) bool {
	if p.Iteration != q.Iteration {
		return p.Iteration < q.Iteration
	}
	return p.Step < q.Step
}

// LoopProposal is the shortcut on the table (CR 732.2a/b).
type LoopProposal struct {
	Count int       `json:"count"`
	End   LoopPoint `json:"end"`
	// ShortenedBy is the last seat that shortened the proposal, if any.
	ShortenedBy uuid.UUID `json:"shortenedBy,omitempty"`
	// Asked is every seat that has answered, in the order they did.
	Asked []uuid.UUID `json:"asked,omitempty"`
}

// Shortened reports whether some seat moved the ending point.
func (p *LoopProposal) Shortened() bool { return p != nil && p.ShortenedBy != uuid.Nil }

// LoopSignature is the part of the board a loop must give back (ADR
// 0193 §3).
type LoopSignature struct {
	Turn     int                  `json:"turn"`
	Step     Step                 `json:"step"`
	Active   uuid.UUID            `json:"active"`
	Priority uuid.UUID            `json:"priority"`
	Live     int                  `json:"live"`
	Pending  int                  `json:"pending,omitempty"`
	Stack    []LoopStackEntry     `json:"stack,omitempty"`
	Board    []LoopPermanentState `json:"board,omitempty"`
}

// LoopStackEntry is one stack item as the signature sees it.
type LoopStackEntry struct {
	Kind       StackItemKind `json:"kind"`
	Source     uuid.UUID     `json:"source,omitempty"`
	Label      string        `json:"label,omitempty"`
	Controller uuid.UUID     `json:"controller,omitempty"`
}

// LoopPermanentState is one permanent as the signature sees it.
type LoopPermanentState struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name,omitempty"`
	Controller uuid.UUID `json:"controller,omitempty"`
	Tapped     bool      `json:"tapped,omitempty"`
	FaceDown   bool      `json:"faceDown,omitempty"`
	PhasedOut  bool      `json:"phasedOut,omitempty"`
	AttachedTo uuid.UUID `json:"attachedTo,omitempty"`
}

// LoopTotals is the allowed-to-differ part of the board, per player.
type LoopTotals struct {
	Players []LoopPlayerTotals `json:"players,omitempty"`
}

// LoopPlayerTotals is one player's allowed-to-differ numbers.
type LoopPlayerTotals struct {
	ID        uuid.UUID      `json:"id"`
	Life      int            `json:"life"`
	Poison    int            `json:"poison,omitempty"`
	Energy    int            `json:"energy,omitempty"`
	Hand      int            `json:"hand,omitempty"`
	Library   int            `json:"library,omitempty"`
	Graveyard int            `json:"graveyard,omitempty"`
	Exile     int            `json:"exile,omitempty"`
	Tokens    int            `json:"tokens,omitempty"`
	Mana      map[string]int `json:"mana,omitempty"`
	Counters  map[string]int `json:"counters,omitempty"`
	// OnPermanents is the counters on the permanents they control, by
	// kind.
	OnPermanents map[string]int `json:"onPermanents,omitempty"`
}

// LoopDelta is one number one iteration changes.
type LoopDelta struct {
	Player uuid.UUID `json:"player"`
	// What names the number: "life", "poison", "energy", "hand",
	// "library", "graveyard", "exile", "tokens", "mana:<symbol>",
	// "counter:<name>" (a player counter) or "permanent:<kind>"
	// (counters on their permanents).
	What   string `json:"what"`
	Amount int    `json:"amount"`
}

// ---------------------------------------------------------------
// Reading the recording
// ---------------------------------------------------------------

// cloneLoopRecording deep-copies a recording.
func cloneLoopRecording(r *LoopRecording) *LoopRecording {
	if r == nil {
		return nil
	}
	out := *r
	out.Start = cloneLoopSignature(r.Start)
	out.Totals = cloneLoopTotals(r.Totals)
	out.StartObjects = append([]uuid.UUID(nil), r.StartObjects...)
	out.Births = append([]LoopObject(nil), r.Births...)
	if r.Steps != nil {
		out.Steps = make([]LoopStep, len(r.Steps))
		for i, s := range r.Steps {
			s.Pending = append([]string(nil), s.Pending...)
			out.Steps[i] = s
		}
	}
	out.Net = append([]LoopDelta(nil), r.Net...)
	if r.Proposal != nil {
		p := *r.Proposal
		p.Asked = append([]uuid.UUID(nil), r.Proposal.Asked...)
		out.Proposal = &p
	}
	return &out
}

func cloneLoopSignature(s LoopSignature) LoopSignature {
	s.Stack = append([]LoopStackEntry(nil), s.Stack...)
	s.Board = append([]LoopPermanentState(nil), s.Board...)
	return s
}

func cloneLoopTotals(t LoopTotals) LoopTotals {
	if t.Players == nil {
		return t
	}
	out := LoopTotals{Players: make([]LoopPlayerTotals, len(t.Players))}
	for i, p := range t.Players {
		p.Mana = cloneIntMap(p.Mana)
		p.Counters = cloneIntMap(p.Counters)
		p.OnPermanents = cloneIntMap(p.OnPermanents)
		out.Players[i] = p
	}
	return out
}

func cloneIntMap(m map[string]int) map[string]int {
	if m == nil {
		return nil
	}
	out := make(map[string]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// CurrentLoopRecording returns a deep copy of the recording, or nil.
func (g *Game) CurrentLoopRecording() *LoopRecording {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return cloneLoopRecording(g.LoopRecording)
}

// LoopRecordingLocked is CurrentLoopRecording for a caller that holds
// the read lock (the view). The pointer is the live one: read only.
func (g *Game) LoopRecordingLocked() *LoopRecording { return g.LoopRecording }

// LoopReplayDue reports whether every seat has answered the proposal,
// so the room must replay it now.
func (g *Game) LoopReplayDue() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.LoopRecording != nil && g.LoopRecording.Status == LoopStatusAccepted
}

// LoopChoicePoints is the steps of r that are seat's own choices: the
// places CR 732.2b lets them shorten the proposal to.
func (r *LoopRecording) LoopChoicePoints(seat uuid.UUID) []int {
	if r == nil {
		return nil
	}
	var out []int
	for i, s := range r.Steps {
		if s.chooser() == seat {
			out = append(out, i)
		}
	}
	return out
}

// chooser is the seat whose choice the step is.
func (s LoopStep) chooser() uuid.UUID {
	if s.Actor != uuid.Nil {
		return s.Actor
	}
	return s.Player
}

// FullEnd is the end of an unshortened proposal of count iterations.
func (r *LoopRecording) FullEnd(count int) LoopPoint {
	return LoopPoint{Iteration: count, Step: len(r.Steps)}
}

// Name is how the log names the loop: the first step's source, and the
// first other source a step touched ("Basalt Monolith ⇄ Rings of
// Brighthearth").
func (r *LoopRecording) Name() string {
	var first, second string
	for _, s := range r.Steps {
		switch {
		case s.Source == "":
		case first == "":
			first = s.Source
		case s.Source != first && second == "":
			second = s.Source
		}
	}
	switch {
	case first == "":
		return "a loop"
	case second == "":
		return first
	}
	return first + " ⇄ " + second
}

// ---------------------------------------------------------------
// Signature and totals
// ---------------------------------------------------------------

// seatIDLocked is the player at seat index i, or uuid.Nil.
func (g *Game) seatIDLocked(i int) uuid.UUID {
	if i < 0 || i >= len(g.Seats) || g.Seats[i] == nil {
		return uuid.Nil
	}
	return g.Seats[i].ID
}

// liveSeatsLocked counts the seats still in the game.
func (g *Game) liveSeatsLocked() int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			n++
		}
	}
	return n
}

// loopSignatureLocked takes the board's loop signature. Caller must
// hold g.mu.
func (g *Game) loopSignatureLocked() LoopSignature {
	s := LoopSignature{
		Turn:     g.Turn.Seq,
		Step:     g.Turn.Step,
		Active:   g.seatIDLocked(g.Turn.ActiveSeat),
		Priority: g.seatIDLocked(g.Turn.PriorityHolder),
		Live:     g.liveSeatsLocked(),
		Pending:  len(g.PendingChoices),
	}
	for _, item := range g.stackItemsBySeqLocked() {
		s.Stack = append(s.Stack, LoopStackEntry{
			Kind:       item.Kind,
			Source:     item.SourceCardID,
			Label:      item.Label,
			Controller: item.Controller,
		})
	}
	add := func(z *Zone, phased bool) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			s.Board = append(s.Board, LoopPermanentState{
				ID:         c.InstanceID,
				Name:       c.Name,
				Controller: c.Controller,
				Tapped:     c.Tapped,
				FaceDown:   c.FaceDown,
				PhasedOut:  phased,
				AttachedTo: c.AttachedTo.ID,
			})
		}
	}
	add(g.Battlefield, false)
	add(g.PhasedOut, true)
	return s
}

// stackItemsBySeqLocked is every stack item, bottom first.
func (g *Game) stackItemsBySeqLocked() []*StackItem {
	items := make([]*StackItem, 0, len(g.StackMeta))
	for _, item := range g.StackMeta {
		if item != nil {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Seq != items[j].Seq {
			return items[i].Seq < items[j].Seq
		}
		return items[i].ID.String() < items[j].ID.String()
	})
	return items
}

// compareLoopSignaturesLocked names the first way now differs from
// start, as a sentence, or returns "" when the loop closed. recorder
// is the seat that must hold priority. Caller must hold g.mu (for the
// player names).
func (g *Game) compareLoopSignaturesLocked(start, now LoopSignature, recorder uuid.UUID) string {
	switch {
	case start.Turn != now.Turn || start.Step != now.Step || start.Active != now.Active:
		return "the turn or step changed"
	case start.Live != now.Live:
		return "a player left the game"
	case now.Pending > 0:
		return "a prompt is still waiting for an answer"
	case now.Priority != recorder:
		return g.seatLabelLocked(recorder) + " does not hold priority"
	}
	if len(start.Stack) != len(now.Stack) {
		return fmt.Sprintf("the stack has %d objects and had %d", len(now.Stack), len(start.Stack))
	}
	for i := range start.Stack {
		a, b := start.Stack[i], now.Stack[i]
		if a.Kind != b.Kind || a.Source != b.Source || a.Label != b.Label || a.Controller != b.Controller {
			return "the stack is not what it was at the start"
		}
	}
	byID := make(map[uuid.UUID]LoopPermanentState, len(now.Board))
	for _, p := range now.Board {
		byID[p.ID] = p
	}
	for _, a := range start.Board {
		b, ok := byID[a.ID]
		name := a.Name
		if name == "" {
			name = "a permanent"
		}
		switch {
		case !ok:
			return name + " is no longer on the battlefield"
		case a.Controller != b.Controller:
			return name + " has a different controller"
		case a.Tapped != b.Tapped:
			if b.Tapped {
				return name + " is tapped at the end and was untapped at the start"
			}
			return name + " is untapped at the end and was tapped at the start"
		case a.FaceDown != b.FaceDown:
			return name + " was turned face up or face down"
		case a.PhasedOut != b.PhasedOut:
			return name + " phased in or out"
		case a.AttachedTo != b.AttachedTo:
			return name + " is attached to something else"
		}
	}
	return ""
}

// loopTotalsLocked counts the allowed-to-differ numbers. Caller must
// hold g.mu.
func (g *Game) loopTotalsLocked() LoopTotals {
	var t LoopTotals
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		pt := LoopPlayerTotals{
			ID:        p.ID,
			Life:      p.Life,
			Poison:    p.Poison,
			Energy:    p.Energy,
			Hand:      p.Hand.Size(),
			Library:   p.Library.Size(),
			Graveyard: p.Graveyard.Size(),
			Counters:  cloneIntMap(p.Counters),
		}
		for _, tok := range p.ManaPool {
			if pt.Mana == nil {
				pt.Mana = map[string]int{}
			}
			pt.Mana[tok.Color] += tok.N()
		}
		if g.Exile != nil {
			for i := range g.Exile.Cards {
				if g.Exile.Cards[i].Owner == p.ID {
					pt.Exile++
				}
			}
		}
		if g.Battlefield != nil {
			for i := range g.Battlefield.Cards {
				c := &g.Battlefield.Cards[i]
				if c.Controller != p.ID {
					continue
				}
				if IsTokenPtr(c) {
					pt.Tokens++
				}
				for k, v := range c.Counters {
					if pt.OnPermanents == nil {
						pt.OnPermanents = map[string]int{}
					}
					pt.OnPermanents[k] += v
				}
			}
		}
		t.Players = append(t.Players, pt)
	}
	return t
}

// loopDeltas is now minus start, in a stable order.
func loopDeltas(start, now LoopTotals) []LoopDelta {
	before := make(map[uuid.UUID]LoopPlayerTotals, len(start.Players))
	for _, p := range start.Players {
		before[p.ID] = p
	}
	var out []LoopDelta
	for _, b := range now.Players {
		a := before[b.ID]
		add := func(what string, d int) {
			if d != 0 {
				out = append(out, LoopDelta{Player: b.ID, What: what, Amount: d})
			}
		}
		add("life", b.Life-a.Life)
		add("poison", b.Poison-a.Poison)
		add("energy", b.Energy-a.Energy)
		add("hand", b.Hand-a.Hand)
		add("library", b.Library-a.Library)
		add("graveyard", b.Graveyard-a.Graveyard)
		add("exile", b.Exile-a.Exile)
		add("tokens", b.Tokens-a.Tokens)
		for _, k := range unionKeys(a.Mana, b.Mana) {
			add("mana:"+k, b.Mana[k]-a.Mana[k])
		}
		for _, k := range unionKeys(a.Counters, b.Counters) {
			add("counter:"+k, b.Counters[k]-a.Counters[k])
		}
		for _, k := range unionKeys(a.OnPermanents, b.OnPermanents) {
			add("permanent:"+k, b.OnPermanents[k]-a.OnPermanents[k])
		}
	}
	return out
}

func unionKeys(a, b map[string]int) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range []map[string]int{a, b} {
		for k := range m {
			if !seen[k] {
				seen[k] = true
				out = append(out, k)
			}
		}
	}
	sort.Strings(out)
	return out
}

// loopNetTextLocked renders deltas, each times `times`, for the log
// and the proposal: "Luke: +3 {C}, library -1; Ann: life -1". Caller
// must hold g.mu.
func (g *Game) loopNetTextLocked(deltas []LoopDelta, times int) string {
	if len(deltas) == 0 {
		return "no net change"
	}
	var parts []string
	var cur uuid.UUID
	var group []string
	flush := func() {
		if len(group) > 0 {
			parts = append(parts, g.seatLabelLocked(cur)+": "+strings.Join(group, ", "))
		}
		group = nil
	}
	for _, d := range deltas {
		if d.Player != cur {
			flush()
			cur = d.Player
		}
		n := signed(d.Amount * times)
		switch {
		case strings.HasPrefix(d.What, "mana:"):
			group = append(group, n+" {"+strings.TrimPrefix(d.What, "mana:")+"}")
		case strings.HasPrefix(d.What, "counter:"):
			group = append(group, strings.TrimPrefix(d.What, "counter:")+" "+n)
		case strings.HasPrefix(d.What, "permanent:"):
			group = append(group, strings.TrimPrefix(d.What, "permanent:")+" counters "+n)
		default:
			group = append(group, d.What+" "+n)
		}
	}
	flush()
	return strings.Join(parts, "; ")
}

func signed(n int) string {
	if n >= 0 {
		return "+" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}

// LoopPublicName is how a recorded step names a player or a card: a
// seat's label, or the name of a card the whole table can see — a
// permanent, or a spell or an ability's source on the stack. A card in
// a hidden zone is not named, so a step's label never says what a
// search found. "" when there is nothing public to say.
func (g *Game) LoopPublicName(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if p := g.playerByIDLocked(id); p != nil {
		return g.seatLabelLocked(id)
	}
	if item := g.StackMeta[id]; item != nil && item.SourceCardID != id {
		id = item.SourceCardID
	}
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut, g.Stack} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			if c := &z.Cards[i]; c.InstanceID == id && !c.FaceDown {
				return c.Name
			}
		}
	}
	return ""
}

// LoopNetText renders the recording's net change for one iteration
// times `times`.
func (g *Game) LoopNetText(r *LoopRecording, times int) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.loopNetTextLocked(r.Net, times)
}

// LoopNetTextLocked is LoopNetText for a caller holding the read lock.
func (g *Game) LoopNetTextLocked(r *LoopRecording, times int) string {
	return g.loopNetTextLocked(r.Net, times)
}

// ---------------------------------------------------------------
// Objects a step can name
// ---------------------------------------------------------------

// loopObjectsLocked lists every object a step can name by an ID that
// did not exist before, in a fixed order: stack items bottom first,
// permanents, phased-out permanents, waiting triggers, pending prompts.
// Caller must hold g.mu.
func (g *Game) loopObjectsLocked() []LoopObject {
	var out []LoopObject
	for _, item := range g.stackItemsBySeqLocked() {
		out = append(out, LoopObject{ID: item.ID, Kind: LoopObjectStack})
	}
	for _, z := range []*Zone{g.Battlefield, g.PhasedOut} {
		if z == nil {
			continue
		}
		for i := range z.Cards {
			out = append(out, LoopObject{ID: z.Cards[i].InstanceID, Kind: LoopObjectPermanent})
		}
	}
	for _, t := range g.PendingTriggers {
		if t != nil {
			out = append(out, LoopObject{ID: t.ID, Kind: LoopObjectTrigger})
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil {
			out = append(out, LoopObject{ID: c.ID, Kind: LoopObjectChoice})
		}
	}
	return out
}

// loopPendingShapeLocked is the open prompts as "kind@seat". Caller
// must hold g.mu.
func (g *Game) loopPendingShapeLocked() []string {
	var out []string
	for _, c := range g.PendingChoices {
		if c != nil {
			out = append(out, string(c.Kind)+"@"+c.Chooser.String())
		}
	}
	return out
}

// LoopPendingShape is the open prompts as "kind@seat", the shape a
// replayed step compares with its recorded Pending.
func (g *Game) LoopPendingShape() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.loopPendingShapeLocked()
}

// LoopBirths tracks the objects created inside one iteration, in the
// order a scan before each step finds them. The recording keeps one on
// the game; a replay keeps one per iteration.
type LoopBirths struct {
	known map[uuid.UUID]struct{}
	list  []LoopObject
}

// NewLoopBirths starts tracking an iteration from the board as it is.
func (g *Game) NewLoopBirths() *LoopBirths {
	g.mu.RLock()
	defer g.mu.RUnlock()
	b := &LoopBirths{known: map[uuid.UUID]struct{}{}}
	for _, o := range g.loopObjectsLocked() {
		b.known[o.ID] = struct{}{}
	}
	return b
}

// Scan adds every object the board has that the tracker has not seen.
func (b *LoopBirths) Scan(g *Game) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	b.scanLocked(g)
}

func (b *LoopBirths) scanLocked(g *Game) {
	for _, o := range g.loopObjectsLocked() {
		if _, ok := b.known[o.ID]; ok {
			continue
		}
		b.known[o.ID] = struct{}{}
		b.list = append(b.list, o)
	}
}

// RefFor is the recorded reference for id, when id was born in the
// iteration.
func (b *LoopBirths) RefFor(id uuid.UUID) (string, bool) {
	if b == nil {
		return "", false
	}
	for k, o := range b.list {
		if o.ID == id {
			return LoopRefPrefix + o.Kind + ":" + strconv.Itoa(k), true
		}
	}
	return "", false
}

// Resolve is the object a recorded reference names in this iteration.
// ok is false when the string is a reference that names nothing here —
// a divergence. isRef is false when the string is not a reference.
func (b *LoopBirths) Resolve(ref string) (id uuid.UUID, isRef, ok bool) {
	if !strings.HasPrefix(ref, LoopRefPrefix) {
		return uuid.Nil, false, false
	}
	rest := strings.TrimPrefix(ref, LoopRefPrefix)
	kind, idx, found := strings.Cut(rest, ":")
	if !found {
		return uuid.Nil, true, false
	}
	k, err := strconv.Atoi(idx)
	if err != nil || k < 0 || b == nil || k >= len(b.list) || b.list[k].Kind != kind {
		return uuid.Nil, true, false
	}
	return b.list[k].ID, true, true
}

// LoopStepDraft is a step being recorded: the births and the open
// prompts as they were just before the action. CommitLoopStep files it
// once the action has succeeded.
type LoopStepDraft struct {
	startSeq uint64
	births   *LoopBirths
	pending  []string
}

// RefFor is the reference a step records for id, when id was created
// inside the iteration.
func (d *LoopStepDraft) RefFor(id uuid.UUID) (string, bool) { return d.births.RefFor(id) }

// BeginLoopStep drafts the step an action is about to be, or returns
// nil when no loop is being recorded.
func (g *Game) BeginLoopStep() *LoopStepDraft {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r := g.LoopRecording
	if r == nil || r.Status != LoopStatusRecording || g.loopReplaying {
		return nil
	}
	b := &LoopBirths{known: make(map[uuid.UUID]struct{}, len(r.StartObjects)+len(r.Births))}
	for _, id := range r.StartObjects {
		b.known[id] = struct{}{}
	}
	for _, o := range r.Births {
		b.known[o.ID] = struct{}{}
	}
	b.list = append([]LoopObject(nil), r.Births...)
	b.scanLocked(g)
	return &LoopStepDraft{startSeq: r.StartSeq, births: b, pending: g.loopPendingShapeLocked()}
}

// CommitLoopStep files a drafted step on the recording it was drafted
// against. A recording that ended in the meantime takes nothing.
func (g *Game) CommitLoopStep(d *LoopStepDraft, step LoopStep) {
	if d == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil || r.Status != LoopStatusRecording || r.StartSeq != d.startSeq || g.loopReplaying {
		return
	}
	r.Births = append([]LoopObject(nil), d.births.list...)
	step.Pending = d.pending
	r.Steps = append(r.Steps, step)
}

// ---------------------------------------------------------------
// The flow
// ---------------------------------------------------------------

// emitLoopLineLocked writes one EventLoopRecording. Caller must hold g.mu.
func (g *Game) emitLoopLineLocked(actor uuid.UUID, call, label string, amount int) {
	g.EmitEvent(Event{Kind: EventLoopRecording, Actor: actor, Call: call, Label: label, Amount: amount})
}

// Loop event phases, EventLoopRecording's Call.
const (
	LoopPhaseArmed     = "armed"
	LoopPhaseStarted   = "started"
	LoopPhaseStopping  = "stopping"
	LoopPhaseReady     = "ready"
	LoopPhaseDiscarded = "discarded"
	LoopPhaseProposed  = "proposed"
	LoopPhaseAccepted  = "accepted"
	LoopPhaseShortened = "shortened"
)

// requireLoopSeatLocked checks player is still in an active game past
// its opening: the condition for asking to record at all. Caller must
// hold g.mu.
func (g *Game) requireLoopSeatLocked(player uuid.UUID) error {
	if g.State != StateActive || g.OpeningRoll != nil || g.MulligansOpen {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(player)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	return nil
}

// requireLoopPriorityLocked checks player may propose a loop right now:
// the game is active, they hold priority and no prompt is open
// (CR 732.2a: "the player with priority may suggest a shortcut").
// Caller must hold g.mu.
func (g *Game) requireLoopPriorityLocked(player uuid.UUID) error {
	if err := g.requireLoopSeatLocked(player); err != nil {
		return err
	}
	if g.seatIDLocked(g.Turn.PriorityHolder) != player {
		return ErrLoopNeedsPriority
	}
	if len(g.PendingChoices) > 0 || len(g.PendingTriggers) > 0 {
		return ErrLoopPromptOpen
	}
	return nil
}

// loopQuietPriorityLocked reports whether player holds priority at a
// point a demonstrated iteration can start or end: the game active and
// past its opening, no prompt open and no trigger waiting to be put on
// the stack. Partway through a resolution there is always a prompt
// open (that is what paused it), so this is never true there. Caller
// must hold g.mu.
func (g *Game) loopQuietPriorityLocked(player uuid.UUID) bool {
	return g.requireLoopPriorityLocked(player) == nil
}

// StartLoopRecording asks to record a loop for player (ADR 0193 §1, CR
// 732.2a). Any seat still in the game may ask, while no other recording
// stands: recording only captures what happens next, and it is the
// proposal, not the recording, that CR 732.2a gives the priority holder.
// When player holds priority with nothing open the recording starts
// now. Otherwise, partway through a resolution or while another seat
// holds priority, it is armed and starts at their next quiet priority,
// so the demonstrated iteration always runs from one priority point of
// theirs to the next (#4736; the closure check needs both ends there).
func (g *Game) StartLoopRecording(player uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.LoopRecording != nil {
		return ErrLoopAlreadyRecording
	}
	if err := g.requireLoopSeatLocked(player); err != nil {
		return err
	}
	r := &LoopRecording{Recorder: player, Status: LoopStatusArmed}
	g.LoopRecording = r
	if g.loopQuietPriorityLocked(player) {
		g.beginLoopRecordingLocked(r)
		return nil
	}
	g.emitLoopLineLocked(player, LoopPhaseArmed, "will record a loop from their next priority (CR 732.2a)", 0)
	return nil
}

// beginLoopRecordingLocked starts capturing an armed recording: the
// start signature, the baseline of the net change and the objects that
// already exist. Caller must hold g.mu, with the recorder at a quiet
// priority.
func (g *Game) beginLoopRecordingLocked(r *LoopRecording) {
	g.emitLoopLineLocked(r.Recorder, LoopPhaseStarted, "started recording a loop (CR 732.2a)", 0)
	objs := g.loopObjectsLocked()
	ids := make([]uuid.UUID, len(objs))
	for i, o := range objs {
		ids[i] = o.ID
	}
	r.Status = LoopStatusRecording
	r.StartSeq = g.eventSeq
	r.Start = g.loopSignatureLocked()
	r.Totals = g.loopTotalsLocked()
	r.StartObjects = ids
}

// StopLoopRecording ends the demonstration. CR 732.2a ends a shortcut's
// sequence where a player has priority, and the closure check compares
// two such points of the recorder's, so the check runs at their quiet
// priority: now if they are there, else at the next one, recording
// until then (#4736). A recording that closes is ready to propose; one
// that does not is discarded, and the log says why.
func (g *Game) StopLoopRecording(player uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil {
		return ErrLoopNoRecording
	}
	if r.Recorder != player {
		return ErrLoopNotRecorder
	}
	if r.Status != LoopStatusRecording {
		return ErrLoopNotRecording
	}
	if len(r.Steps) == 0 {
		return ErrLoopEmpty
	}
	if !g.loopQuietPriorityLocked(player) {
		if !r.StopAtPriority {
			r.StopAtPriority = true
			g.emitLoopLineLocked(player, LoopPhaseStopping,
				fmt.Sprintf("will stop recording the loop at their next priority (%d steps so far)", len(r.Steps)), len(r.Steps))
		}
		return nil
	}
	g.closeLoopRecordingLocked(r)
	return nil
}

// closeLoopRecordingLocked runs the closure check (ADR 0193 §3) with
// the recorder at a quiet priority: the recording is ready, or it is
// discarded with the first difference in the log. Caller must hold g.mu.
func (g *Game) closeLoopRecordingLocked(r *LoopRecording) {
	r.StopAtPriority = false
	if why := g.compareLoopSignaturesLocked(r.Start, g.loopSignatureLocked(), r.Recorder); why != "" {
		g.discardLoopRecordingLocked("the loop does not close: " + why)
		return
	}
	r.Status = LoopStatusReady
	r.Net = loopDeltas(r.Totals, g.loopTotalsLocked())
	g.emitLoopLineLocked(r.Recorder, LoopPhaseReady,
		fmt.Sprintf("recorded a loop of %d steps (each time: %s)", len(r.Steps), g.loopNetTextLocked(r.Net, 1)), len(r.Steps))
}

// CancelLoopRecording discards the recording and withdraws any
// outstanding proposal. Only the recorder may.
func (g *Game) CancelLoopRecording(player uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil {
		return ErrLoopNoRecording
	}
	if r.Recorder != player {
		return ErrLoopNotRecorder
	}
	g.discardLoopRecordingLocked("they cancelled it")
	return nil
}

// DiscardLoopRecording ends any recording, with reason in the log. The
// room calls it when an undo is taken (ADR 0193 §2 and owner decision 8).
func (g *Game) DiscardLoopRecording(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.LoopRecording == nil {
		return
	}
	g.discardLoopRecordingLocked(reason)
}

// discardLoopRecordingLocked drops the recording and its prompts and
// writes the line. Caller must hold g.mu.
func (g *Game) discardLoopRecordingLocked(reason string) {
	r := g.LoopRecording
	if r == nil {
		return
	}
	g.LoopRecording = nil
	g.dropLoopProposalPromptsLocked()
	g.emitLoopLineLocked(r.Recorder, LoopPhaseDiscarded, "stopped recording the loop: "+reason, 0)
}

// dropLoopProposalPromptsLocked withdraws every loop_proposal prompt.
// Caller must hold g.mu.
func (g *Game) dropLoopProposalPromptsLocked() {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == PendingChoiceLoopProposal {
			g.dropChoiceLocked(i)
		}
	}
}

// ProposeLoop is CR 732.2a's proposal: the recorded iteration, count
// more times. Each other seat is then asked in turn order.
func (g *Game) ProposeLoop(player uuid.UUID, count int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil {
		return ErrLoopNoRecording
	}
	if r.Recorder != player {
		return ErrLoopNotRecorder
	}
	if r.Status != LoopStatusReady {
		return ErrLoopNotReady
	}
	if count < 1 || count > MaxLoopShortcutIterations {
		return ErrLoopCount
	}
	if err := g.requireLoopPriorityLocked(player); err != nil {
		return err
	}
	if g.compareLoopSignaturesLocked(r.Start, g.loopSignatureLocked(), r.Recorder) != "" {
		return ErrLoopBoardChanged
	}
	r.Status = LoopStatusProposed
	r.Proposal = &LoopProposal{Count: count, End: r.FullEnd(count)}
	g.emitLoopLineLocked(player, LoopPhaseProposed,
		fmt.Sprintf("proposes repeating %s %d more times (CR 732.2a)", r.Name(), count), count)
	g.askNextLoopSeatLocked()
	return nil
}

// askNextLoopSeatLocked queues the proposal to the next seat in turn
// order after the recorder that has not answered and is still in the
// game, or marks the proposal accepted when there is none. Caller must
// hold g.mu.
func (g *Game) askNextLoopSeatLocked() {
	r := g.LoopRecording
	if r == nil || r.Status != LoopStatusProposed || r.Proposal == nil {
		return
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceLoopProposal {
			return
		}
	}
	from := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == r.Recorder {
			from = i
		}
	}
	asked := map[uuid.UUID]bool{}
	for _, id := range r.Proposal.Asked {
		asked[id] = true
	}
	n := len(g.Seats)
	for off := 1; off < n; off++ {
		p := g.Seats[(from+off+n)%n]
		if p == nil || p.Eliminated || asked[p.ID] || p.ID == r.Recorder {
			continue
		}
		g.QueueChoiceForEffect(PendingChoice{
			Kind:    PendingChoiceLoopProposal,
			Chooser: p.ID,
			Count:   1,
			Reason:  "Loop shortcut (CR 732.2b)",
		})
		return
	}
	r.Status = LoopStatusAccepted
}

// ResolveLoopProposal answers a loop_proposal prompt: accept when
// shorten is nil, else shorten the proposal to that point, which must be
// one of the chooser's own choices and no later than the current end
// (CR 732.2b).
func (g *Game) ResolveLoopProposal(choiceID, chooser uuid.UUID, shorten *LoopPoint) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceLoopProposal {
		return ErrInvalidParam
	}
	if choice.Chooser != chooser {
		return ErrNotTheChooser
	}
	r := g.LoopRecording
	if r == nil || r.Status != LoopStatusProposed || r.Proposal == nil {
		return ErrLoopNoRecording
	}
	if shorten != nil {
		pt := *shorten
		if pt.Iteration < 1 || pt.Iteration > r.Proposal.Count || pt.Step < 0 || pt.Step >= len(r.Steps) ||
			r.Steps[pt.Step].chooser() != chooser {
			return ErrLoopShortenPlace
		}
		if r.Proposal.End.before(pt) {
			return ErrLoopShortenTooLate
		}
		r.Proposal.End = pt
		r.Proposal.ShortenedBy = chooser
		g.emitLoopLineLocked(chooser, LoopPhaseShortened,
			fmt.Sprintf("shortens the loop to iteration %d, before their step %d (CR 732.2b)", pt.Iteration, pt.Step+1), pt.Iteration)
	} else {
		g.emitLoopLineLocked(chooser, LoopPhaseAccepted, "accepts the loop shortcut (CR 732.2b)", 0)
	}
	r.Proposal.Asked = append(r.Proposal.Asked, chooser)
	g.dequeueChoiceLocked(idx)
	g.askNextLoopSeatLocked()
	return nil
}

// TendLoopRecording moves a recording along with the game: an armed one
// starts, and a stopping one is checked, once the recorder holds
// priority with nothing open (#4736); one the game has moved past — the
// game ended, a player left, the step or turn changed — ends; and the
// proposal is re-asked when the seat it was waiting on has left.
// actions.Dispatch and the room call it after every commit; it is cheap
// when there is no recording.
func (g *Game) TendLoopRecording() {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil || g.loopReplaying || r.Status == LoopStatusAccepted {
		return
	}
	switch {
	case g.State != StateActive:
		g.discardLoopRecordingLocked("the game ended")
		return
	}
	if p := g.playerByIDLocked(r.Recorder); p == nil || p.Eliminated {
		g.discardLoopRecordingLocked("they left the game")
		return
	}
	if r.Status == LoopStatusArmed {
		// Nothing is captured yet, so the step and the seats may move
		// on: the recording begins wherever the recorder next holds
		// priority with nothing open.
		if g.loopQuietPriorityLocked(r.Recorder) {
			g.beginLoopRecordingLocked(r)
		}
		return
	}
	if g.Turn.Seq != r.Start.Turn || g.Turn.Step != r.Start.Step || g.seatIDLocked(g.Turn.ActiveSeat) != r.Start.Active {
		g.discardLoopRecordingLocked("the step changed")
		return
	}
	if g.liveSeatsLocked() != r.Start.Live {
		g.discardLoopRecordingLocked("a player left the game")
		return
	}
	if r.Status == LoopStatusRecording && r.StopAtPriority && g.loopQuietPriorityLocked(r.Recorder) {
		g.closeLoopRecordingLocked(r)
		return
	}
	g.askNextLoopSeatLocked()
}

// ---------------------------------------------------------------
// Replay
// ---------------------------------------------------------------

// BeginLoopReplay starts replaying the accepted proposal: the breaker
// sits out from here until FinishLoopReplay, and the log opens the run
// it will fold into one line. Returns a copy of the recording.
func (g *Game) BeginLoopReplay() (*LoopRecording, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := g.LoopRecording
	if r == nil || r.Status != LoopStatusAccepted || r.Proposal == nil {
		return nil, ErrLoopNoRecording
	}
	g.loopReplaying = true
	g.EmitEvent(Event{Kind: EventLoopReplayBegin, Actor: r.Recorder})
	return cloneLoopRecording(r), nil
}

// LoopReplaying reports whether a replay is running.
func (g *Game) LoopReplaying() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.loopReplaying
}

// LoopIterationGap compares the board with the recording's start
// signature after a replayed iteration: "" when it is the same loop
// again, else the first difference.
func (g *Game) LoopIterationGap(r *LoopRecording) string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.compareLoopSignaturesLocked(r.Start, g.loopSignatureLocked(), r.Recorder)
}

// LoopReplayOutcome is how a replay ended, for its one log line.
type LoopReplayOutcome struct {
	// Completed is how many whole iterations stand.
	Completed int
	// Partial is set when a shortened proposal's last, partial
	// iteration was reached.
	Partial bool
	// Stopped is the divergence that stopped the replay, or "".
	Stopped string
}

// FinishLoopReplay ends a replay: the breaker comes back with every run
// cleared (as a decision clears them, ADR 0193 §5), the recording is
// spent, and the replay's one public line is written.
func (g *Game) FinishLoopReplay(r *LoopRecording, out LoopReplayOutcome) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loopReplaying = false
	g.LoopRecording = nil
	g.dropLoopProposalPromptsLocked()
	g.notePlayerDecisionLocked()
	count := 0
	if r.Proposal != nil {
		count = r.Proposal.Count
	}
	line := fmt.Sprintf("repeats %s ×%d", r.Name(), out.Completed)
	if out.Completed > 0 {
		line += ": " + g.loopNetTextLocked(r.Net, out.Completed)
	}
	if r.Proposal.Shortened() && out.Stopped == "" {
		line += fmt.Sprintf(" (shortened by %s to iteration %d, before their step %d)",
			g.seatLabelLocked(r.Proposal.ShortenedBy), r.Proposal.End.Iteration, r.Proposal.End.Step+1)
	}
	if out.Stopped != "" {
		line += fmt.Sprintf(" (stopped after %d of %d: %s)", out.Completed, count, out.Stopped)
	}
	g.EmitEvent(Event{Kind: EventLoopReplayed, Actor: r.Recorder, Label: line, Amount: out.Completed})
}
