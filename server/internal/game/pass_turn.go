package game

import "github.com/google/uuid"

// pass_turn.go — the active player's "Pass turn" (#2881).
//
// Pass turn used to jump the cursor straight to the next seat's untap
// step. Every step left in the turn was skipped outright, so "at the
// beginning of combat", "at the beginning of your end step" and every
// other step trigger never fired, and the cleanup discard never
// happened. CR 500.1 has every phase happen every turn, even if
// nothing happens in it, and CR 500.6 has an "at the beginning of"
// trigger fire as its step begins.
//
// Pass turn is now a standing instruction to pass priority for the
// active player, every time they would hold it, until their turn ends.
// The steps are walked by the one priority engine (passPriorityLocked),
// so each step begins, its turn-based actions happen, its triggers go on
// the stack the next time a player would receive priority (CR 117.5,
// 603.3), and everything on the stack resolves only after every player
// has passed in succession (CR 117.4).
//
// Only the active player is passed for. Every other player still
// receives priority in each step (CR 117.3d) and passes or acts as they
// would without it: their own client's automatic passing, or their bot
// runner, decides, so pass turn never makes them miss a window.
//
// The passing stops, and the player gets the decision, whenever:
//
//   - a prompt the table must wait for is open (a "may", a target, an
//     ordering, the cleanup discard — ChoiceBlocksTable), or any prompt
//     at all is owed by the passing player;
//   - the CR 732 loop breaker has suspended automatic passing;
//   - the pass itself is refused (an attack requirement, CR 508.1d:
//     "attacks each combat if able" is not waved off by passing).
//
// It resumes on its own once that is settled: the dispatcher calls
// SettlePassTurn after every action, and the room after every
// automatic answer. It ends when the turn does, by any route.
//
// Carried by Clone and RestoreFrom, so undoing the pass_turn takes the
// instruction back with it. Not part of a restore point: a server that
// restarts in the middle of a passed turn leaves the active player
// holding priority, and they press Pass turn again.

// passTurnOrder is the standing pass: who asked, and in which turn.
type passTurnOrder struct {
	Player uuid.UUID
	Seq    int
}

// clonePassTurnOrder copies the order, so a clone never shares it.
func clonePassTurnOrder(o *passTurnOrder) *passTurnOrder {
	if o == nil {
		return nil
	}
	c := *o
	return &c
}

// maxPassTurnPasses bounds one settle. A turn with nothing in it costs
// about one pass per step; the rest is one per resolution of the
// player's own priority. Hitting it leaves the order standing, and the
// next action settles again.
const maxPassTurnPasses = 512

// PassTurn is the active player's "pass turn": pass priority for them
// whenever they hold it until the turn ends, walking every remaining
// step (CR 500.1). See the file comment.
//
// Refused while a prompt the table waits for is open (#730,
// choice_gate.go): it is answered first, as before.
func (g *Game) PassTurn() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if c := g.blockingChoiceLocked(); c != nil {
		return choicePendingErrorLocked(c)
	}
	active := g.Turn.ActiveSeat
	if active < 0 || active >= len(g.Seats) || g.Seats[active] == nil {
		return ErrPlayerNotFound
	}
	g.passTurn = &passTurnOrder{Player: g.Seats[active].ID, Seq: g.Turn.Seq}
	g.settlePassTurnLocked()
	return nil
}

// PassingTurn reports whether the active player has passed the turn
// and the passing is still standing.
func (g *Game) PassingTurn() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.passTurnLiveLocked()
}

// SettlePassTurn passes priority for an active player who passed the
// turn, for as long as they hold it and nothing stops the passing. A
// no-op without a standing pass. actions.Dispatch calls it after every
// action, beside SettleResolution, and the room after each automatic
// answer.
func (g *Game) SettlePassTurn() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.settlePassTurnLocked()
}

// passTurnLiveLocked drops a pass whose turn is over and reports
// whether one still stands.
//
// Caller must hold g.mu (a read lock only reports).
func (g *Game) passTurnLiveLocked() bool {
	o := g.passTurn
	if o == nil {
		return false
	}
	if g.State != StateActive || g.Turn.Seq != o.Seq {
		return false
	}
	a := g.Turn.ActiveSeat
	return a >= 0 && a < len(g.Seats) && g.Seats[a] != nil && g.Seats[a].ID == o.Player && !g.Seats[a].Eliminated
}

// settlePassTurnLocked is SettlePassTurn under the caller's lock.
//
// Caller must hold g.mu in write mode.
func (g *Game) settlePassTurnLocked() {
	for i := 0; i < maxPassTurnPasses; i++ {
		if !g.passTurnLiveLocked() {
			g.passTurn = nil
			return
		}
		if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
			return
		}
		if g.blockingChoiceLocked() != nil || g.LoopNotice != nil || g.choiceOwedByLocked(g.passTurn.Player) {
			return
		}
		if err := g.passPriorityAsHolderLocked(); err != nil {
			return
		}
	}
}

// choiceOwedByLocked reports whether any prompt, blocking or not, is
// addressed to the player.
//
// Caller must hold g.mu.
func (g *Game) choiceOwedByLocked(player uuid.UUID) bool {
	for i := range g.PendingChoices {
		if g.PendingChoices[i].Chooser == player {
			return true
		}
	}
	return false
}

// EndTurnNowForTest is the old sandbox pass_turn: it ends the turn at
// once and starts the next seat's, without the steps in between (no
// end step, so no "at the beginning of the end step" triggers, and no
// cleanup discard). Combat ends and the cleanup sweep runs through the
// rotation seam (rotation.go, #766). Tests use it to reach the next
// turn without walking this one; no player-facing path reaches it.
func (g *Game) EndTurnNowForTest() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if c := g.blockingChoiceLocked(); c != nil {
		return choicePendingErrorLocked(c)
	}
	g.clearCombatLocked()
	g.sweepTurnEndLocked()
	g.beginNextTurnLocked()
	g.runStepEntryHooksLocked()
	g.runStateChecksLocked()
	return nil
}
