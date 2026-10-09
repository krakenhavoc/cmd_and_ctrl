package botarena

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// opening.go is the arena's mulligan numbers (#2693): how often a seat
// mulligans, the size of the hand it keeps, and whether it made its
// land drop on its own second, third and fourth turns. A mulligan rule
// is judged by the last: a hand kept on its land count that then stops
// drawing lands misses those drops.
//
// Like turnmana.go, it reads only what the runner's observer is handed:
// the seat's own filtered view and the move it dispatched.

// landDropFirst and landDropLast are the seat's own turns whose land
// drop is counted: its second through its fourth.
const (
	landDropFirst = 2
	landDropLast  = 4
)

// Opening is one seat's (or one contestant's) mulligan numbers.
type Opening struct {
	// Seats is the seat-games that kept a hand.
	Seats int `json:"seats"`
	// Mulligans is the mulligans those seats took.
	Mulligans int `json:"mulligans"`
	// Kept counts the kept hands by size: Kept[7] is seven-card keeps.
	Kept map[int]int `json:"kept,omitempty"`
	// DropTurns is the seat's own turns 2–4 it reached (made a decision
	// in), and Missed is those in which it played no land.
	DropTurns int `json:"drop_turns"`
	Missed    int `json:"missed_drops"`
}

func (o *Opening) add(x Opening) {
	o.Seats += x.Seats
	o.Mulligans += x.Mulligans
	o.DropTurns += x.DropTurns
	o.Missed += x.Missed
	for k, n := range x.Kept {
		if o.Kept == nil {
			o.Kept = map[int]int{}
		}
		o.Kept[k] += n
	}
}

// MulligansPerGame is Mulligans / Seats, 0 with no seats.
func (o Opening) MulligansPerGame() float64 {
	if o.Seats == 0 {
		return 0
	}
	return float64(o.Mulligans) / float64(o.Seats)
}

// openingWatch is one seat's observer for the numbers.
type openingWatch struct {
	mu        sync.Mutex
	seat      uuid.UUID
	mulligans int
	kept      int // hand size kept, -1 before the keep
	// turns is the own turns seen, in order, by Turn.Seq; land marks
	// those in which a land was played.
	turns []int
	land  map[int]bool
}

func newOpeningWatch(seat uuid.UUID) *openingWatch {
	return &openingWatch{seat: seat, kept: -1, land: map[int]bool{}}
}

// Observe implements aiseat.DecisionObserver.
func (w *openingWatch) Observe(ev aiseat.DecisionEvent) {
	if ev.Seat != w.seat {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	v := &ev.Input.View
	m, ok := dispatchedMove(ev)
	if ok && m.Kind == legal.KindMulligan {
		switch m.Type {
		case legal.TypeMulligan:
			w.mulligans++
		case legal.TypeKeepHand:
			w.kept = ownHandSize(v, w.seat.String())
		}
		return
	}
	as := v.Turn.ActiveSeat
	if as < 0 || as >= len(v.Seats) || v.Seats[as].ID != w.seat.String() {
		return
	}
	seq := v.Turn.Seq
	if n := len(w.turns); n == 0 || w.turns[n-1] != seq {
		w.turns = append(w.turns, seq)
	}
	if ok && m.Kind == legal.KindLand {
		w.land[seq] = true
	}
}

func ownHandSize(v *protocol.GameView, seat string) int {
	for i := range v.Seats {
		if v.Seats[i].ID == seat {
			return len(v.Seats[i].Hand.Cards)
		}
	}
	return 0
}

// result is the seat's numbers for the game.
func (w *openingWatch) result() Opening {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := Opening{Mulligans: w.mulligans}
	if w.kept >= 0 {
		out.Seats = 1
		out.Kept = map[int]int{w.kept: 1}
	}
	for i, seq := range w.turns {
		if own := i + 1; own < landDropFirst || own > landDropLast {
			continue
		}
		out.DropTurns++
		if !w.land[seq] {
			out.Missed++
		}
	}
	return out
}

func writeOpening(b *strings.Builder, s Summary) {
	if len(s.PerContestant) == 0 && len(s.PerPolicy) == 0 {
		return
	}
	b.WriteString("\n### Opening hands (#2693)\n\n")
	b.WriteString("| policy | deck | keeps | mulligans | per game | kept 7 / 6 / other | own turns 2–4 | missed land drops | missed % |\n")
	b.WriteString("|---|---|---:|---:|---:|---|---:|---:|---:|\n")
	row := func(policy, deck string, o Opening) {
		other := 0
		sizes := make([]int, 0, len(o.Kept))
		for k := range o.Kept {
			sizes = append(sizes, k)
		}
		sort.Ints(sizes)
		for _, k := range sizes {
			if k != 7 && k != 6 {
				other += o.Kept[k]
			}
		}
		fmt.Fprintf(b, "| %s | %s | %d | %d | %.2f | %d / %d / %d | %d | %d | %s |\n",
			policy, deck, o.Seats, o.Mulligans, o.MulligansPerGame(), o.Kept[7], o.Kept[6], other,
			o.DropTurns, o.Missed, pct(ratio(o.Missed, o.DropTurns)))
	}
	for _, n := range s.Policies() {
		row(n, "all", s.PerPolicy[n].Opening)
	}
	for _, t := range s.PerContestant {
		row(t.Policy, orDash(t.Deck), t.Opening)
	}
	b.WriteString("\n*`keeps` is the seat-games that kept a hand, and `per game` is mulligans per keep. `own turns 2–4` is a seat's own second to fourth turns that it reached; a `missed land drop` is one in which it played no land (a land put onto the battlefield by a spell is not a land drop).*\n")
}
