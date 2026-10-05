package game

import (
	"errors"

	"github.com/google/uuid"
)

// table_roll.go is ADR 0121 §5: the "Roll a die" table action. A
// player rolls a d6 or a d20, or flips a coin, for fun, at any time
// while the game is active, the opening roll and the mulligan
// included.
//
// It is NOT a game roll. No effect instructed it (CR 706.1), so it
// emits no EventRollDie or EventFlipCoin, and no "whenever you roll one
// or more dice" or "whenever you win a coin flip" ability can see it.
// It emits an EventTableRoll, which no watcher matches. It changes
// nothing but the log, never opens a pending choice, and mints no undo
// entry (actions.MintsNoUndo), so nobody waits for anybody.
//
// Three things keep it out of the game's randomness and its history:
//
//   - its own stream, drawn by tableRollDrawLocked outside the turn
//     counters, so no card's draw moves;
//   - Game.tableRollNext, carried forward across an undo, so a roll
//     after an undo is a fresh one;
//   - the ring below, which RestoreFrom re-emits, so an undo of
//     somebody's earlier action does not erase the line.

// The three things a table roll can be.
const (
	TableDieD6   = "d6"
	TableDieD20  = "d20"
	TableDieCoin = "coin"
)

// tableRollRingMax is how many recent table rolls an undo can carry
// across (ADR 0121 §5). The hub lets a seat roll once per 2 s, so 32
// covers any undo a table can reach.
const tableRollRingMax = 32

// ErrUnknownTableDie refuses a roll_table_die that is not a d6, a d20
// or a coin.
var ErrUnknownTableDie = errors.New("game: a table roll is a d6, a d20 or a coin")

// RollTableDie rolls a d6 or a d20, or flips a coin, at the table for
// playerID (ADR 0121 §5) and returns the event it emitted. die is
// TableDieD6, TableDieD20 or TableDieCoin. Takes the write lock.
func (g *Game) RollTableDie(playerID uuid.UUID, die string) (Event, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return Event{}, ErrGameNotActive
	}
	sides := 0
	switch die {
	case TableDieD6:
		sides = 6
	case TableDieD20:
		sides = 20
	case TableDieCoin:
	default:
		return Event{}, ErrUnknownTableDie
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return Event{}, ErrPlayerNotFound
	}
	if p.Eliminated {
		return Event{}, ErrPlayerEliminated
	}
	ev := Event{Kind: EventTableRoll, Actor: playerID, Sides: sides}
	if sides > 0 {
		n, id := g.tableRollDrawLocked(playerID, sides)
		ev.Amount, ev.RollID = n+1, id
	} else {
		// true is heads, as FlipCoinsForEffect reads it.
		n, id := g.tableRollDrawLocked(playerID, 2)
		ev.Label, ev.RollID = "tails", id
		if n == 0 {
			ev.Label = "heads"
		}
	}
	return g.emitTableRollLocked(ev), nil
}

// emitTableRollLocked emits ev and remembers it, as emitted, in the
// ring. Returns the stamped event. Caller must hold g.mu for writing.
func (g *Game) emitTableRollLocked(ev Event) Event {
	g.EmitEvent(ev)
	stamped := g.Events[len(g.Events)-1]
	for i := range g.tableRolls {
		if g.tableRolls[i].RollID == stamped.RollID {
			g.tableRolls[i] = stamped
			return stamped
		}
	}
	g.tableRolls = append(g.tableRolls, stamped)
	if over := len(g.tableRolls) - tableRollRingMax; over > 0 {
		g.tableRolls = append([]Event(nil), g.tableRolls[over:]...)
	}
	return stamped
}

// reemitTableRollsLocked puts back the table rolls an undo just cut off
// the log (ADR 0121 §5). RestoreFrom truncates Events to the restored
// game's length; every ring entry whose Seq is past that end is emitted
// again, oldest first, with its payload unchanged. It gets a new Seq
// and keeps its RollID, so a client that keys the roll on its RollID
// does not animate it twice. The line moves to after the restored
// history, which is still after everything that happened before it.
//
// A ring entry at or before the restored end is still in the log: the
// undo stack only ever restores an ancestor of the live history, and
// the ring holds each roll's latest Seq. Caller must hold g.mu for
// writing.
func (g *Game) reemitTableRollsLocked() {
	if len(g.tableRolls) == 0 {
		return
	}
	end := g.eventSeq
	var cut []Event
	for _, ev := range g.tableRolls {
		if ev.Seq > end {
			cut = append(cut, ev)
		}
	}
	for _, ev := range cut {
		// EmitEvent stamps a fresh Seq and the open batch.
		g.emitTableRollLocked(ev)
	}
}

// SetTriggerOrderPreference sets playerID's "always ask me to order my
// triggers" preference (#1530). It is a seat setting, not a play: it
// emits no event and mints no undo entry (actions.MintsNoUndo), and
// RestoreFrom carries it across an undo. Takes the write lock.
func (g *Game) SetTriggerOrderPreference(playerID uuid.UUID, alwaysAsk bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	p.TriggerOrderAlwaysAsk = alwaysAsk
	return nil
}
