package actions

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// table_verb_gates.go — who may send the table-wide sandbox verbs (#4131,
// ADR 0028's amendment of 2026-10-19).
//
// Before #4131 these verbs took no caller or ownership check, on the
// "sandbox among friends" reading: any seat could announce a trigger as
// another seat, take damage off another seat's creature, edit another
// seat's commander damage, goad on another's behalf or end a vote it
// did not call. Owner decision of 2026-10-19: gate the risky ones, so
// that only the AFFECTED player, or the controller of the AFFECTED
// card, can use them, and no table-wide override survives beyond the
// admin's (Caller uuid.Nil), which every gate below keeps for repairs.
//
// The rule each verb follows:
//
//	announce_trigger     as yourself (playerScopedActions), with a source you control
//	set_commander_damage only on yourself as the damaged player
//	mark_damage          the card's controller, whichever way the delta points
//	set_goaded           the creature's controller, or the goading player as themself
//	set_promise          either end of the promise
//	set_monarch          take it yourself, or hand it on / clear it while you hold it
//	set_initiative       the same
//	advance_step         the active player or the priority holder
//	clear_combat         the active player or the priority holder
//	start_vote           as yourself (playerScopedActions)
//	end_vote             the player who called the vote
//	counter_spell/ability  unchanged: the priority holder (see Dispatch)

// ErrNotTurnOrPriority is returned when advance_step or clear_combat is
// sent by a seat that is neither the active player nor the priority
// holder. Admin / spectator callers bypass it.
var ErrNotTurnOrPriority = errors.New("actions: caller is neither the active player nor the priority holder")

// ErrNotVoteInitiator is returned when end_vote is sent by a seat other
// than the one that called the vote.
var ErrNotVoteInitiator = errors.New("actions: only the player who called the vote may end it")

// ErrMarkerNotYours is returned when a seat tries to give the monarchy
// or the initiative to another seat, or to clear a marker it does not
// hold. A seat takes a marker for itself; only its holder hands it on.
var ErrMarkerNotYours = errors.New("actions: take the marker yourself, or release it while you hold it")

// requireActiveOrPriority (called with overrideCaller, so a seated admin keeps
// the repair) passes the active player and the priority
// holder, which is how passing works: whoever may move the game on may
// also move the cursor or wipe a wedged combat. uuid.Nil (admin /
// spectator) bypasses. While priority is parked for a block declaration
// nobody holds it, and the active player is then the one who may act.
func requireActiveOrPriority(g *game.Game, caller uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	snap := g.Snapshot()
	if snap.State != game.StateActive {
		return nil
	}
	if i := snap.Turn.ActiveSeat; i >= 0 && i < len(snap.Seats) {
		if s := snap.Seats[i]; s != nil && s.ID == caller {
			return nil
		}
	}
	if i := snap.Turn.PriorityHolder; i >= 0 && i < len(snap.Seats) {
		if s := snap.Seats[i]; s != nil && s.ID == caller {
			return nil
		}
	}
	return ErrNotTurnOrPriority
}

// requireMarkerChange gates set_monarch and set_initiative. `to` is the
// player taking the marker, uuid.Nil to clear it; `holder` is who holds
// it now. A seat may take it for itself, or release or hand it on while
// it is the holder. The admin bypasses.
func requireMarkerChange(caller, to, holder uuid.UUID) error {
	if caller == uuid.Nil || to == caller || holder == caller {
		return nil
	}
	return ErrMarkerNotYours
}

// requireVoteInitiator gates end_vote: the player who called the vote
// ends it. A vote whose initiator has left the game (eliminated, or no
// longer seated) would otherwise sit open for ever, so any seat may
// close that one. No open vote passes through so EndVote reports
// ErrNoVote.
func requireVoteInitiator(g *game.Game, caller uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	allowed := true
	g.ReadSnapshot(func() {
		if g.Vote == nil || g.Vote.Initiator == caller {
			return
		}
		// Someone else called it: refused, unless the caller of the
		// vote has left the table.
		gone := true
		for _, p := range g.Seats {
			if p != nil && p.ID == g.Vote.Initiator {
				gone = p.Eliminated
				break
			}
		}
		allowed = gone
	})
	if allowed {
		return nil
	}
	return ErrNotVoteInitiator
}

// holderOfMarker reads the current monarch or initiative holder.
func holderOfMarker(g *game.Game, initiative bool) uuid.UUID {
	var holder uuid.UUID
	g.ReadSnapshot(func() {
		if initiative {
			holder = g.Initiative
		} else {
			holder = g.Monarch
		}
	})
	return holder
}

// requireTriggerSource gates announce_trigger's source (#4131): a
// permanent's trigger is its controller's. A source that has left the
// battlefield also admits its owner and the player who controlled it as
// it last existed there (CR 603.3a, CR 603.10a), so the player whose
// stolen creature died can still announce its dies trigger. A third
// seat is refused. uuid.Nil (admin) bypasses, and an unknown card passes
// through to the mutation's ErrCardNotFound.
func requireTriggerSource(g *game.Game, caller, sourceID uuid.UUID) error {
	if caller == uuid.Nil {
		return nil
	}
	claimants, ok := g.TriggerSourceClaimants(sourceID)
	if !ok {
		return nil
	}
	for _, c := range claimants {
		if c == caller {
			return nil
		}
	}
	return game.ErrCardCallerMismatch
}
