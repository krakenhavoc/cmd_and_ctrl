package game

import "github.com/google/uuid"

// priority_succession.go — CR 117.4's "all players pass in succession"
// as engine state (#2275, ADR 0007's 2026-10-05 amendment).
//
//	117.4. If all players pass in succession (that is, if all players
//	pass without taking any actions in between passing), the spell or
//	ability on top of the stack resolves or, if the stack is empty,
//	the phase or step ends.
//
// The gap. passPriorityLocked used to move priority to the next seat
// and treat arriving back at the ACTIVE seat as "everyone passed". That
// is the same thing only when the round of passes began with the active
// player. A non-active player who casts a spell keeps priority (CR
// 117.3c), so the round restarts with them; when they passed, the walk
// reached the active seat and resolved the spell without the active
// player — or any seat between them in turn order — ever holding
// priority over it. Prod game 497de7d2: a Stroke of Genius cast in its
// opponent's draw step resolved through a Counterspell and a Mana Drain.
//
// The state. Turn.PassedInSuccession is the set of seats that have
// passed since the succession last began. A pass adds the passer; the
// round is over when every seat still in the game is in the set
// (allPassedInSuccessionLocked). Priority still moves to the next
// player in turn order on every other pass (CR 117.3d).
//
// What begins a new succession — every writer is in this file, or is
// one of the calls to it named here:
//
//   - Priority granted at a boundary (grantPriorityLocked): a step
//     begins (popTurnPlanLocked, a new Turn), the top of the stack
//     resolves (CR 117.3b), a combat declaration or the block
//     declaration ends (CR 117.3a), the cleanup step grants priority
//     (CR 514.3a), and a holder who left the game hands it on.
//   - Anything put on the stack (nextStackSeqLocked, which every stack
//     push stamps its order through): a cast, an activation, a trigger,
//     a copy. A trigger is not an action a player took, but the object
//     on top is a new one, and every seat gets to see it before it
//     resolves — the conservative reading, and the one that never lets
//     a new object resolve on passes made before it existed.
//   - The priority holder taking an action that uses no stack (CR
//     117.3c, noteActionTakenLocked): a land play (CR 116.2a), another
//     special action (PerformSpecialAction), a mana ability, a manual
//     loyalty activation, and the sandbox's counter verbs.
//   - A player leaving the game (CR 800.4a takes their objects off the
//     stack, so what the others passed over is not what is there now).
//
// A seat that has left is not waited for: the round is over when every
// seat still in the game has passed, so a departure never strands one.

// grantPriorityLocked gives `seat` priority at a boundary — CR 117.3a
// (a step begins, a turn-based action ends) or CR 117.3b (a spell or
// ability resolved) — and begins a new succession of passes with it.
// NoPriority parks the cursor (untap, cleanup, a block declaration)
// and clears the succession just the same.
//
// Caller must hold g.mu in write mode.
func (g *Game) grantPriorityLocked(seat int) {
	g.Turn.PriorityHolder = seat
	g.Turn.PassedInSuccession = 0
}

// restartPassSuccessionLocked forgets every pass made so far: something
// happened that the seats who passed did not pass over. Priority stays
// where it is.
//
// Caller must hold g.mu in write mode.
func (g *Game) restartPassSuccessionLocked() {
	g.Turn.PassedInSuccession = 0
}

// noteActionTakenLocked is CR 117.3c: a player who holds priority and
// takes an action — casts, activates, plays a land, takes a special
// action, activates a mana ability — keeps priority, and the passes
// made before it no longer count. A player who acts WITHOUT holding
// priority (tapping mana to answer a "pay {1}" prompt) has taken no
// action in the CR 117.4 sense, and the succession stands.
//
// Caller must hold g.mu in write mode.
func (g *Game) noteActionTakenLocked(player uuid.UUID) {
	ph := g.Turn.PriorityHolder
	if ph < 0 || ph >= len(g.Seats) || g.Seats[ph] == nil || g.Seats[ph].ID != player {
		return
	}
	g.restartPassSuccessionLocked()
}

// notePassLocked records that `seat` passed priority.
//
// Caller must hold g.mu in write mode.
func (g *Game) notePassLocked(seat int) {
	g.Turn.PassedInSuccession = g.Turn.PassedInSuccession.With(seat)
}

// allPassedInSuccessionLocked reports whether every seat still in the
// game has passed in succession — CR 117.4's condition for the top of
// the stack to resolve, or the step to end. Seats that have left are
// not waited for (CR 800.4a).
//
// Caller must hold g.mu.
func (g *Game) allPassedInSuccessionLocked() bool {
	for i, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if !g.Turn.PassedInSuccession.Has(i) {
			return false
		}
	}
	return true
}
