package botarena

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// turnmana.go is ADR 0136 §8's two arena numbers, the measurement the
// turn plan is judged by:
//
//   - Stranded mana. In each of a seat's own turns, the mana it could
//     still make at its last pass in its last main phase, and whether
//     a cast was on offer then. A turn is STRANDED when that pass left
//     StrandedMana or more mana unspent with a cast still on offer.
//     Acceptance bar P2 compares the stranded share of `heuristic`
//     with `heuristic-noplan`'s.
//   - Plan misses. A window in which the previous window's plan
//     (aiseat.Trace.Plan) named a next member that is not offered now,
//     though no other seat acted in between. P6 holds them under 5% of
//     planned windows. It says how good the plan's mana model is: a
//     miss is a plan that thought a second cast was payable and was
//     wrong.
//
// Both come from the runner's observer, like the Cards section, so they
// need no decision log, and both read only what the observer is
// handed: each seat's own filtered view, its move list and its trace.
//
// "The mana it could still make" is counted from the view the way ADR
// 0136 §2's mana model counts what the bot has: the floating pool, plus
// each untapped permanent the seat controls with a repeatable mana
// ability it can activate now (a {T} ability with no sacrifice or exile
// cost; not a summoning-sick creature, CR 302.6; not a row the view
// greys out), at the most that one activation nets. It is a count, not
// a colour check: a turn that ended with two red open and only a blue
// spell in hand is not stranded, because the blue spell was not on
// offer.

// StrandedMana is how much unspent mana makes a turn stranded (ADR
// 0136 §8): two or more.
const StrandedMana = 2

// TurnMana is one seat's (or one contestant's) ADR 0136 §8 numbers.
// The zero value is a seat that never reached a main-phase pass.
type TurnMana struct {
	// Turns is the seat's own turns in which it passed at least once in
	// a main phase with an empty stack: the turns stranded mana is
	// measured over.
	Turns int `json:"turns"`
	// Stranded is the turns whose last such pass left StrandedMana or
	// more mana unspent with a cast on offer.
	Stranded int `json:"stranded"`
	// Unspent is the mana the seat could still make at those last
	// passes, summed over Turns. Unspent / Turns is the mean.
	Unspent int `json:"unspent"`
	// Idle is the turns whose last pass left StrandedMana or more mana
	// unspent, cast on offer or not: Stranded's context, the mana a hand
	// with nothing castable left on the table.
	Idle int `json:"idle"`
	// Planned is the windows whose trace carried a plan of two or more
	// members (aiseat.Trace.Plan): the plan-miss denominator.
	Planned int `json:"planned_windows"`
	// Checked is the windows in which a plan's next member was looked
	// for: the seat's next main-phase window with an empty stack after
	// it made the plan's first move, in the same phase, with no other
	// seat having acted in between.
	Checked int `json:"plan_checked"`
	// Misses is the Checked windows in which that member was not
	// offered.
	Misses int `json:"plan_misses"`
}

// StrandedShare is Stranded / Turns, 0 with no turns.
func (t TurnMana) StrandedShare() float64 { return ratio(t.Stranded, t.Turns) }

// MissShare is Misses / Planned, ADR 0136's P6 ("under 5% of planned
// windows"), 0 with no plans.
func (t TurnMana) MissShare() float64 { return ratio(t.Misses, t.Planned) }

// MeanUnspent is Unspent / Turns, 0 with no turns.
func (t TurnMana) MeanUnspent() float64 {
	if t.Turns == 0 {
		return 0
	}
	return float64(t.Unspent) / float64(t.Turns)
}

func (t *TurnMana) add(o TurnMana) {
	t.Turns += o.Turns
	t.Stranded += o.Stranded
	t.Unspent += o.Unspent
	t.Idle += o.Idle
	t.Planned += o.Planned
	t.Checked += o.Checked
	t.Misses += o.Misses
}

func ratio(n, d int) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

// turnManaWatch is one game's observer for the two numbers. It is
// shared by every seat at the table, because a plan miss only counts
// when no OTHER seat acted between the two windows, and only a
// table-wide observer sees the other seats' moves.
//
// In a concurrent run the seats' events arrive in the order their
// goroutines reach the observer, which is the order they were decided
// in up to the scheduler; in a lockstep run the order is exact.
type turnManaWatch struct {
	mu    sync.Mutex
	seats map[uuid.UUID]*turnManaSeat
}

type turnManaSeat struct {
	// last is each own turn's latest main-phase pass, by Turn.Seq.
	last    map[int]lastPass
	planned int
	checked int
	misses  int
	pending *planNext
}

// lastPass is what one main-phase pass left.
type lastPass struct {
	mana int
	cast bool
}

// planNext is the member a plan expects to be offered next.
type planNext struct {
	source      uuid.UUID
	turn, phase int
	step        string
	// interrupted is set when another seat made a move other than a
	// pass after the plan's first move: whatever that move was, a miss
	// after it is not the mana model's.
	interrupted bool
}

func newTurnManaWatch() *turnManaWatch {
	return &turnManaWatch{seats: map[uuid.UUID]*turnManaSeat{}}
}

func (w *turnManaWatch) seat(id uuid.UUID) *turnManaSeat {
	s := w.seats[id]
	if s == nil {
		s = &turnManaSeat{last: map[int]lastPass{}}
		w.seats[id] = s
	}
	return s
}

// Observe implements aiseat.DecisionObserver.
func (w *turnManaWatch) Observe(ev aiseat.DecisionEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()

	dispatched, ok := dispatchedMove(ev)
	if ok && dispatched.Kind != legal.KindPass {
		for id, s := range w.seats {
			if id != ev.Seat && s.pending != nil {
				s.pending.interrupted = true
			}
		}
	}

	v := &ev.Input.View
	s := w.seat(ev.Seat)
	main := ownMainPhaseEmptyStack(v, ev.Seat)

	if p := s.pending; p != nil && main {
		s.pending = nil
		if p.turn == v.Turn.Seq && p.phase == v.Turn.PhaseID && p.step == v.Turn.Step && !p.interrupted {
			s.checked++
			if !offersCast(ev.Input.Moves, p.source) {
				s.misses++
			}
		}
	}

	if plan := ev.Trace.Plan; len(plan) >= 2 {
		s.planned++
		if ok && ev.Index == plan[0].Index && main {
			for _, m := range plan[1:] {
				if m.Held || m.Index < 0 || m.Index >= len(ev.Input.Moves) {
					continue
				}
				s.pending = &planNext{source: ev.Input.Moves[m.Index].Source, turn: v.Turn.Seq, phase: v.Turn.PhaseID, step: v.Turn.Step}
				break
			}
		}
	}

	if main && ok && dispatched.Kind == legal.KindPass {
		s.last[v.Turn.Seq] = lastPass{mana: manaLeft(v, ev.Seat.String()), cast: offersAnyCast(ev.Input.Moves)}
	}
}

// forSeat is one seat's numbers for the game so far.
func (w *turnManaWatch) forSeat(id uuid.UUID) TurnMana {
	w.mu.Lock()
	defer w.mu.Unlock()
	s := w.seats[id]
	if s == nil {
		return TurnMana{}
	}
	out := TurnMana{Turns: len(s.last), Planned: s.planned, Checked: s.checked, Misses: s.misses}
	for _, p := range s.last {
		out.Unspent += p.mana
		if p.mana >= StrandedMana {
			out.Idle++
			if p.cast {
				out.Stranded++
			}
		}
	}
	return out
}

// dispatchedMove is the move the runner dispatched and the engine
// applied in this window.
func dispatchedMove(ev aiseat.DecisionEvent) (legal.Move, bool) {
	if !ev.Applied || ev.Index < 0 || ev.Index >= len(ev.Input.Moves) {
		return legal.Move{}, false
	}
	return ev.Input.Moves[ev.Index], true
}

// ownMainPhaseEmptyStack is the heuristic's sorcery-speed window, read
// the same way (heuristic.go's state): the seat's own turn, a main
// phase, and nothing on the stack (CR 307.1; an ability shows up only
// in stack_items, #1352).
func ownMainPhaseEmptyStack(v *protocol.GameView, seat uuid.UUID) bool {
	as := v.Turn.ActiveSeat
	if as < 0 || as >= len(v.Seats) || v.Seats[as].ID != seat.String() {
		return false
	}
	if len(v.Stack.Cards) > 0 || len(v.StackItems) > 0 {
		return false
	}
	return v.Turn.Step == "precombat_main" || v.Turn.Step == "postcombat_main"
}

func offersCast(moves []legal.Move, source uuid.UUID) bool {
	for _, m := range moves {
		if m.Kind == legal.KindCast && m.Source == source {
			return true
		}
	}
	return false
}

func offersAnyCast(moves []legal.Move) bool {
	for _, m := range moves {
		if m.Kind == legal.KindCast {
			return true
		}
	}
	return false
}

// manaLeft is the mana the seat could still make: its floating pool
// plus each untapped permanent it controls at what one activation of
// its best repeatable mana ability nets.
func manaLeft(v *protocol.GameView, seat string) int {
	n := 0
	for i := range v.Seats {
		if v.Seats[i].ID == seat {
			n += len(v.Seats[i].ManaPool)
		}
	}
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		if c.Controller != seat || c.Tapped || c.SummoningSick {
			continue
		}
		n += sourceMana(c)
	}
	return n
}

// sourceMana is what one untapped permanent can add now: the best of
// its repeatable mana abilities, net of a filter's own mana cost (a
// Signet's {1} in, {W}{U} out, nets one). A land with no ability rows
// on the view counts one, as the heuristic's repeatableMana does.
func sourceMana(c *protocol.CardView) int {
	best := 0
	for i := range c.ManaAbilities {
		ab := &c.ManaAbilities[i]
		if !ab.TapCost || ab.SacrificeCost || ab.ExileSelf || ab.SacrificeLabel != "" || ab.ExilePermanentLabel != "" ||
			ab.AddsNoMana || ab.ConditionUnmet || ab.Exhausted || ab.CantActivate != "" {
			continue
		}
		if n := producedAmount(ab.Produced) - costAmount(ab.ManaCost); n > best {
			best = n
		}
	}
	if best == 0 && len(c.ManaAbilities) == 0 && isLand(c.TypeLine) {
		return 1
	}
	return best
}

// producedAmount counts the mana one activation adds from its
// `produced` string, as the heuristic's manaAmount does: each brace is
// one mana, a choice "{W|U}" is one, a counted choice "{W3|U3}" is
// three, and an empty string (an output the view cannot size) is one.
func producedAmount(produced string) int {
	if produced == "" {
		return 1
	}
	total := 0
	for _, sym := range symbols(produced) {
		if k := strings.IndexByte(sym, '|'); k >= 0 {
			sym = sym[:k]
		}
		n := 1
		if k := strings.IndexAny(sym, "0123456789"); k > 0 {
			if m, err := strconv.Atoi(sym[k:]); err == nil && m > 0 {
				n = m
			}
		}
		total += n
	}
	if total == 0 {
		return 1
	}
	return total
}

// costAmount is a mana cost's total: a number is that much, any other
// symbol one.
func costAmount(cost string) int {
	total := 0
	for _, sym := range symbols(cost) {
		if n, err := strconv.Atoi(sym); err == nil {
			total += n
			continue
		}
		total++
	}
	return total
}

// symbols splits "{2}{G}" into "2", "G".
func symbols(s string) []string {
	var out []string
	for {
		i := strings.IndexByte(s, '{')
		if i < 0 {
			return out
		}
		j := strings.IndexByte(s[i:], '}')
		if j < 0 {
			return out
		}
		if sym := s[i+1 : i+j]; sym != "" {
			out = append(out, sym)
		}
		s = s[i+j+1:]
	}
}

// writeTurnMana renders the Turn mana section.
func writeTurnMana(b *strings.Builder, s Summary) {
	if len(s.PerContestant) == 0 && len(s.PerPolicy) == 0 {
		return
	}
	b.WriteString("\n### Turn mana (ADR 0136 §8)\n\n")
	b.WriteString("| policy | deck | own turns | stranded | stranded % | idle | mean unspent | planned windows | checked | plan misses | miss % of planned |\n")
	b.WriteString("|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|\n")
	row := func(policy, deck string, t TurnMana) {
		fmt.Fprintf(b, "| %s | %s | %d | %d | %s | %d | %.2f | %d | %d | %d | %s |\n",
			policy, deck, t.Turns, t.Stranded, pct(t.StrandedShare()), t.Idle, t.MeanUnspent(),
			t.Planned, t.Checked, t.Misses, pct(t.MissShare()))
	}
	for _, n := range s.Policies() {
		row(n, "all", s.PerPolicy[n].TurnMana)
	}
	for _, t := range s.PerContestant {
		row(t.Policy, orDash(t.Deck), t.TurnMana)
	}
	fmt.Fprintf(b, "\n*`own turns` are the turns a seat passed in its own main phase with an empty stack; the numbers are read at its last such pass. `stranded` is a turn that ended with %d or more mana it could still make and a cast on offer; `idle` is %d or more mana with or without one; `mean unspent` is the mana left at that pass, averaged over own turns. `planned windows` carried a turn plan of two or more casts; `checked` is the seat's next main-phase window after making a plan's first move with no other seat acting in between, and a `plan miss` is one where the plan's next cast was not offered. No plan is made before ADR 0136 PR 4.*\n", StrandedMana, StrandedMana)
}
