package game

import (
	"errors"
	"math/rand/v2"

	"github.com/google/uuid"
)

// opening_roll.go is ADR 0121 §1–§2: the opening roll as a pre-game
// window. CR 103.1: "the players determine which one of them will
// choose who takes the first turn … (flipping a coin, rolling dice,
// etc.)". Every seat rolls a d20, tied leaders roll again, and the one
// leader left — the chooser — chooses who takes the first turn. Only
// then is anything shuffled or drawn (CR 103.3, 103.5, 903.7), so
// nobody chooses with a hand in view.
//
// The window is StateActive with MulligansOpen true, as at Start, plus
// a non-nil Game.OpeningRoll. MulligansOpen already parks the turn
// machinery (step hooks, instant windows, the untap choice), so the
// window needs no new State. The cursor waits at the pre-game turn:
// Turn.Seq 0, untap step, nobody holding priority.
//
// THE SAME WINNER EITHER WAY. A seat's k-th opening die is the k-th
// draw on its ("roll", seat, nil) stream at turn index 0. It does not
// depend on when the seat rolls, on who rolled first, or on whether the
// host pressed the button for it, so the interactive roll and the
// automatic one (StartWithFirstPlayerRoll) find the same chooser for
// the same key. The shuffle stays at turn index 0 on its own stream,
// so every library comes out as it did when the deal came first.

const startingPlayerDieSides = 20

// OpeningRoll is the open opening-roll window (ADR 0121 §1). Round 1
// is every seat; each later round is the tied leaders of the one
// before. Pure data, embedded by value in the snapshot.
type OpeningRoll struct {
	Rounds []OpeningRollRound `json:"rounds"`
	// Chooser is the seat that chooses who takes the first turn, or -1
	// while the roll is still going.
	Chooser int `json:"chooser"`
}

// OpeningRollRound is one round of the opening roll: the seats that
// roll in it, in seat order, and their dice in the order rolled.
type OpeningRollRound struct {
	Seats []int            `json:"seats"`
	Rolls []OpeningRollDie `json:"rolls,omitempty"`
}

// OpeningRollDie is one opening die. By is the seat that pressed the
// button: the roller itself, or the host who rolled for everyone left
// (§3). OpeningRollByAdmin when the server admin, seated nowhere,
// pressed it.
type OpeningRollDie struct {
	Seat   int `json:"seat"`
	Result int `json:"result"`
	By     int `json:"by"`
}

// OpeningRollByAdmin is OpeningRollDie.By for a die the server admin
// rolled from no seat.
const OpeningRollByAdmin = -1

// The three Labels EventOpeningRoll carries.
const (
	// OpeningRollTie — a round ended with several seats on the high
	// result; Event.Seats are those seats, who roll again, and
	// Event.Amount the result they tied on.
	OpeningRollTie = "tie"
	// OpeningRollWon — one leader is left; Actor is that seat's player,
	// the chooser, and Event.Amount its winning result (0 when it won
	// because every other contender left before it rolled).
	OpeningRollWon = "won"
	// OpeningRollRolledFor — the host pressed "Roll for everyone left";
	// Actor is the host (uuid.Nil for the server admin) and Event.Seats
	// the seats it rolled for, its own excluded.
	OpeningRollRolledFor = "rolled_for"
)

var (
	// ErrOpeningRollOpen refuses every action but the opening roll's
	// own while the roll is open (ADR 0121 §1). KeepHand and Mulligan
	// return it too: no hand exists yet.
	ErrOpeningRollOpen = errors.New("game: the opening roll is not finished")
	// ErrNoOpeningRoll is returned by the opening roll's verbs when no
	// roll is open.
	ErrNoOpeningRoll = errors.New("game: there is no opening roll in progress")
	// ErrNotYourRoll is returned when a seat that is not in the current
	// round, or has already rolled in it, asks to roll.
	ErrNotYourRoll = errors.New("game: you have no die to roll in this round of the opening roll")
	// ErrNobodyLeftToRoll is returned by HostRollRemaining when every
	// seat in the current round has rolled.
	ErrNobodyLeftToRoll = errors.New("game: everyone in this round of the opening roll has rolled")
	// ErrNoChooserYet is returned by ChooseStartingPlayer while the roll
	// has no single winner.
	ErrNoChooserYet = errors.New("game: the opening roll has no winner yet")
	// ErrNotChooser is returned when anyone but the winner of the
	// opening roll tries to choose who takes the first turn (CR 103.1).
	ErrNotChooser = errors.New("game: only the winner of the opening roll chooses who takes the first turn")
)

// StartWithOpeningRoll is Start for a table that rolls for the first
// turn itself (ADR 0121 §1): life totals and undo budgets are set as
// Start sets them, and the opening roll opens. Nothing is shuffled or
// dealt until the winner chooses (ChooseStartingPlayer). The RNG
// argument seeds the game's key exactly as Start's does.
func (g *Game) StartWithOpeningRoll(r *rand.Rand) error {
	if r == nil {
		return g.start(nil, startOpeningRoll)
	}
	key := rngKeyFrom(r)
	return g.start(&key, startOpeningRoll)
}

// OpeningRollOpen reports whether the opening roll is open. Takes the
// read lock.
func (g *Game) OpeningRollOpen() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.OpeningRoll != nil
}

// RollOpening rolls the player's die in the current round (ADR 0121
// §3, roll_opening). Refused with ErrNotYourRoll unless the seat is in
// the round and has not rolled in it.
func (g *Game) RollOpening(playerID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.OpeningRoll == nil {
		return ErrNoOpeningRoll
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Eliminated {
		return ErrPlayerEliminated
	}
	if !g.owesOpeningDieLocked(p.Seat) {
		return ErrNotYourRoll
	}
	g.rollOpeningDieLocked(p.Seat, p.Seat)
	return nil
}

// HostRollRemaining rolls, in seat order, for every seat in the current
// round that has not rolled (ADR 0121 §3, host_roll_remaining). by is
// the player who pressed the button, uuid.Nil for the server admin.
// WHO may press it is the WebSocket edge's question (Room.CanManageTable),
// as for set_table_settings. A tie it causes opens a new round with its
// own dice to roll.
func (g *Game) HostRollRemaining(by uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.OpeningRoll == nil {
		return ErrNoOpeningRoll
	}
	bySeat := OpeningRollByAdmin
	if by != uuid.Nil {
		p := g.playerByIDLocked(by)
		if p == nil {
			return ErrPlayerNotFound
		}
		bySeat = p.Seat
	}
	round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
	var missing, others []int
	for _, seat := range round.Seats {
		if g.owesOpeningDieLocked(seat) {
			missing = append(missing, seat)
			if seat != bySeat {
				others = append(others, seat)
			}
		}
	}
	if len(missing) == 0 {
		return ErrNobodyLeftToRoll
	}
	if len(others) > 0 {
		g.EmitEvent(Event{Kind: EventOpeningRoll, Label: OpeningRollRolledFor, Actor: by, Seats: others})
	}
	for _, seat := range missing {
		g.rollOpeningDieLocked(seat, bySeat)
	}
	return nil
}

// ChooseStartingPlayer is the chooser's choice (ADR 0121 §2,
// choose_starting_player; CR 103.1): seat takes the first turn. Any
// seat still in the game may be chosen, the chooser's own included.
// Then the libraries are shuffled and the opening hands dealt
// (CR 103.3, 103.5, 903.7), the window closes, and the mulligan opens
// as it always has. There is no undo for it.
func (g *Game) ChooseStartingPlayer(chooser uuid.UUID, seat int) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.OpeningRoll == nil {
		return ErrNoOpeningRoll
	}
	if g.OpeningRoll.Chooser < 0 {
		return ErrNoChooserYet
	}
	p := g.playerByIDLocked(chooser)
	if p == nil {
		return ErrPlayerNotFound
	}
	if p.Seat != g.OpeningRoll.Chooser {
		return ErrNotChooser
	}
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return ErrInvalidParam
	}
	if g.Seats[seat].Eliminated {
		return ErrPlayerEliminated
	}
	g.chooseStartingSeatLocked(seat)
	return nil
}

// openOpeningRollLocked opens the window: round 1 is every seat still
// in the game, and the cursor parks at the pre-game turn.
func (g *Game) openOpeningRollLocked() {
	seats := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			seats = append(seats, p.Seat)
		}
	}
	g.OpeningRoll = &OpeningRoll{Rounds: []OpeningRollRound{{Seats: seats}}, Chooser: -1}
	g.Turn = Turn{
		Step:           StepUntap,
		Phase:          PhaseOf(StepUntap),
		PhaseID:        templatePhaseID(StepUntap),
		PriorityHolder: NoPriority,
	}
}

// rollOpeningAutomaticallyLocked is StartWithFirstPlayerRoll's half of
// the window: every contender rolls in seat order, round by round,
// until one leader is left, and the chooser takes the first turn
// themselves — the pre-0121 behaviour, on the same streams.
func (g *Game) rollOpeningAutomaticallyLocked() {
	for g.OpeningRoll != nil && g.OpeningRoll.Chooser < 0 {
		round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
		for _, seat := range round.Seats {
			if g.owesOpeningDieLocked(seat) {
				g.rollOpeningDieLocked(seat, seat)
			}
		}
	}
	if g.OpeningRoll != nil {
		g.chooseStartingSeatLocked(g.OpeningRoll.Chooser)
	}
}

// owesOpeningDieLocked reports whether seat is in the current round
// and has not rolled in it.
func (g *Game) owesOpeningDieLocked(seat int) bool {
	if g.OpeningRoll == nil || len(g.OpeningRoll.Rounds) == 0 {
		return false
	}
	round := g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
	in := false
	for _, s := range round.Seats {
		if s == seat {
			in = true
			break
		}
	}
	if !in {
		return false
	}
	for _, d := range round.Rolls {
		if d.Seat == seat {
			return false
		}
	}
	return true
}

// rollOpeningDieLocked rolls seat's next opening d20 into the current
// round, through the ordinary random-effect path (one public
// EventRollDie on the ("roll", seat, nil) stream), and settles the
// round once every seat in it has rolled.
func (g *Game) rollOpeningDieLocked(seat, by int) {
	rolls, _ := g.RollDiceForEffect(RandomDraw{Player: g.Seats[seat].ID}, startingPlayerDieSides, 1)
	round := &g.OpeningRoll.Rounds[len(g.OpeningRoll.Rounds)-1]
	round.Rolls = append(round.Rolls, OpeningRollDie{Seat: seat, Result: rolls[0], By: by})
	if len(round.Rolls) == len(round.Seats) {
		g.settleOpeningRollLocked()
	}
}

// settleOpeningRollLocked reads the standings off the latest round
// that still has a seat in it. A single seat left there is the chooser.
// A round every seat has rolled in ends: one leader is the chooser,
// several roll again in a new round. Anything else is still rolling.
// Called when a round's last die lands and when a seat leaves (§1), so
// it emits only on a change: a new round, or a different chooser.
func (g *Game) settleOpeningRollLocked() {
	or := g.OpeningRoll
	for len(or.Rounds) > 0 && len(or.Rounds[len(or.Rounds)-1].Seats) == 0 {
		or.Rounds = or.Rounds[:len(or.Rounds)-1]
	}
	if len(or.Rounds) == 0 {
		// Nobody is left to roll; the departure that did it has ended
		// the game (checkGameOverLocked).
		or.Chooser = -1
		return
	}
	round := or.Rounds[len(or.Rounds)-1]
	if len(round.Seats) == 1 {
		g.setOpeningChooserLocked(round.Seats[0], g.latestOpeningResultLocked(round.Seats[0]))
		return
	}
	if len(round.Rolls) < len(round.Seats) {
		or.Chooser = -1
		return
	}
	leaders, high := openingRoundLeaders(round)
	if len(leaders) == 1 {
		g.setOpeningChooserLocked(leaders[0], high)
		return
	}
	or.Chooser = -1
	or.Rounds = append(or.Rounds, OpeningRollRound{Seats: leaders})
	g.EmitEvent(Event{Kind: EventOpeningRoll, Label: OpeningRollTie, Seats: append([]int(nil), leaders...), Amount: high})
}

// openingRoundLeaders returns the seats on the round's high result, in
// seat order, and that result. Pure, so the tie rule can be proved
// without searching for a key that produces each sequence.
func openingRoundLeaders(round OpeningRollRound) ([]int, int) {
	result := make(map[int]int, len(round.Rolls))
	for _, d := range round.Rolls {
		result[d.Seat] = d.Result
	}
	high := 0
	var leaders []int
	for _, seat := range round.Seats {
		switch r := result[seat]; {
		case r > high:
			high = r
			leaders = append(leaders[:0], seat)
		case r == high:
			leaders = append(leaders, seat)
		}
	}
	return leaders, high
}

// setOpeningChooserLocked records the chooser and announces it once.
func (g *Game) setOpeningChooserLocked(seat, result int) {
	if g.OpeningRoll.Chooser == seat {
		return
	}
	g.OpeningRoll.Chooser = seat
	g.EmitEvent(Event{Kind: EventOpeningRoll, Label: OpeningRollWon, Actor: g.Seats[seat].ID, Amount: result})
}

// latestOpeningResultLocked is seat's most recent opening die, 0 if it
// has none.
func (g *Game) latestOpeningResultLocked(seat int) int {
	for i := len(g.OpeningRoll.Rounds) - 1; i >= 0; i-- {
		rolls := g.OpeningRoll.Rounds[i].Rolls
		for j := len(rolls) - 1; j >= 0; j-- {
			if rolls[j].Seat == seat {
				return rolls[j].Result
			}
		}
	}
	return 0
}

// leaveOpeningRollLocked is a concession during the roll (ADR 0121 §1).
// There is no active player yet (CR 103.8 has not happened), so the
// seat leaves without the active-player departure path: it is struck
// from every round and the standings are read again, the chooser's
// included. A departure that leaves one seat ends the game as any
// other does.
func (g *Game) leaveOpeningRollLocked(p *Player) {
	if !g.leaveGameLocked(p, LossConcede, uuid.Nil) {
		return
	}
	if g.checkGameOverLocked() {
		return
	}
	or := g.OpeningRoll
	for i := range or.Rounds {
		round := &or.Rounds[i]
		round.Seats = removeSeat(round.Seats, p.Seat)
		kept := round.Rolls[:0]
		for _, d := range round.Rolls {
			if d.Seat != p.Seat {
				kept = append(kept, d)
			}
		}
		round.Rolls = kept
	}
	if or.Chooser == p.Seat {
		or.Chooser = -1
	}
	g.settleOpeningRollLocked()
}

func removeSeat(seats []int, seat int) []int {
	out := seats[:0]
	for _, s := range seats {
		if s != seat {
			out = append(out, s)
		}
	}
	return out
}

// chooseStartingSeatLocked closes the window with seat taking the first
// turn: the choice is announced, every library is shuffled and every
// opening hand dealt, and the cursor moves to turn 1.
func (g *Game) chooseStartingSeatLocked(seat int) {
	chooser := g.Seats[g.OpeningRoll.Chooser].ID
	g.EmitEvent(Event{Kind: EventStartingPlayer, Actor: chooser, Target: g.Seats[seat].ID})
	g.OpeningRoll = nil
	g.dealAndBeginLocked(seat)
}

// dealAndBeginLocked is the end of the pregame procedure (CR 103.3,
// 103.5, 903.7): each player still in the game shuffles and draws
// seven, then the starting player's turn 1 begins. The shuffle runs
// BEFORE the cursor moves, so it is the pre-game turn index's first
// draw on each seat's shuffle stream — the same draw it was when the
// deal came first.
func (g *Game) dealAndBeginLocked(startingSeat int) {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		p.Library.Shuffle(g.randForLocked(rngStream{kind: rngStreamShuffle, player: p.ID}))
		// Deal an opening hand of 7. If the library is too short to
		// satisfy 7 (a malformed deck), stop early — the partial hand
		// is still valid and tests can use small decks.
		for i := 0; i < OpeningHandSize; i++ {
			c, err := p.Library.PopTop()
			if err != nil {
				break
			}
			p.Hand.PushTop(c)
		}
		// S13.5: opening-hand cards are known to their owner only.
		// Library cards have no knowers (post-shuffle order is
		// unknown to everyone).
		for i := range p.Hand.Cards {
			p.Hand.Cards[i].AddKnower(p.ID)
		}
	}
	g.StartingSeat = startingSeat
	g.Turn = newStartingTurn(startingSeat)
	// The one turn that does not begin through the rotation seam
	// still counts as a turn begun (ADR 0063 Decision 3).
	g.noteTurnBegunLocked(startingSeat)
}

// cloneOpeningRoll deep-copies the window, so an undo entry or a
// snapshot never shares the live rounds.
func cloneOpeningRoll(or *OpeningRoll) *OpeningRoll {
	if or == nil {
		return nil
	}
	out := &OpeningRoll{Chooser: or.Chooser}
	if or.Rounds == nil {
		return out
	}
	out.Rounds = make([]OpeningRollRound, len(or.Rounds))
	for i, r := range or.Rounds {
		out.Rounds[i] = OpeningRollRound{
			Seats: append([]int(nil), r.Seats...),
			Rolls: append([]OpeningRollDie(nil), r.Rolls...),
		}
	}
	return out
}
