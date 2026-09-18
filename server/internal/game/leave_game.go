package game

import "github.com/google/uuid"

// leave_game.go is CR 800.4a: what happens to a departed player's
// OBJECTS when they concede or lose. The rest of the departure — the
// seat flag, the stack cleanup, the game-over check and the turn
// rotation — lives in mutations.go and rotation.go; this file is only
// the part the rule spells out object by object.
//
// The rule, in its order (CR 800.4a, Aug 7 2026):
//
//  1. every object the departing player OWNS leaves the game;
//  2. effects that gave that player control of anything END;
//  3. non-card objects they controlled on the stack cease to exist;
//  4. anything STILL controlled by them — another player's permanent
//     they had taken by an effect that did not end at step 2 — is
//     exiled.
//
// It is explicitly NOT a state-based action: it happens the moment the
// player leaves, not at the next priority boundary. So it runs inside
// leaveGameLocked, which is the one door both Concede and the SBA loss
// loop go through.
//
// Step 3 is older than this file (`cleanupStackForEliminatedLocked`,
// S13.1) and stays where it is; leaveGameObjectsLocked runs after it,
// which is also what folds the spell CARDS that step used to exile
// into step 1's removal — a card owned by the player who left leaves
// the game, it does not sit in exile.
//
// ADR 0060 has the decisions: why removal is a real removal rather
// than a "left the game" bucket, why nothing here fires a leaves-the-
// battlefield or dies trigger, and what happens to their commander.

// survivingSeatsLocked counts the seats still in the game. Caller must
// hold g.mu.
func (g *Game) survivingSeatsLocked() int {
	n := 0
	for _, s := range g.Seats {
		if !s.Eliminated {
			n++
		}
	}
	return n
}

// leaveGameObjectsLocked performs steps 1, 2 and 4 of CR 800.4a for
// the player who has just left, in that order — the order matters,
// and is the difference between a Mind Control that leaves with its
// controller (the creature goes back to its owner) and one that does
// not (the creature would be exiled).
//
// Caller must hold g.mu, and must already have set p.Eliminated.
func (g *Game) leaveGameObjectsLocked(playerID uuid.UUID) {
	// 1. Everything they own leaves the game.
	removed := g.removeObjectsOwnedByLocked(playerID)

	// 2. Their control effects end. Every control-changing effect the
	//    engine has is a layer-2 continuous effect sourced from a
	//    permanent (ADR 0012 / the S24 Mind Control work), so step 1
	//    has already taken the SOURCE away and all that is owed is a
	//    recompute: `printedCharacteristic` seeds layer 2 from
	//    `BaseController`, so the creature goes back to whoever
	//    controlled it when it entered. The bump is by hand because
	//    nothing emitted an event for the removal above — deliberately,
	//    see removeObjectsOwnedByLocked — and the recompute is
	//    unconditional because step 4 reads Card.Controller off the
	//    back of it.
	g.layerVersion.Add(1)
	g.recomputeLayersLocked()

	// 4. Anything still under their control belongs to somebody else
	//    and has nobody to control it. Exiled.
	g.exileStillControlledLocked(playerID)

	// CR 800.4b/d, and the same reasoning cleanupStackForEliminatedLocked
	// gives for PendingTriggers: nothing goes on the stack under the
	// control of a player who is not in the game. A delayed trigger
	// they scheduled has no controller to put it on the stack for, and
	// the elimination event itself may have harvested a trigger of
	// theirs a moment ago.
	g.dropDelayedTriggersForLocked(playerID)
	g.dropPendingTriggersForLocked(playerID)

	if removed == 0 {
		return
	}
	// Two prompt families can be left asking about objects that are no
	// longer there. Both are the same re-checks executeBattlefieldLeaveLocked
	// runs after any other permanent leaves the battlefield, for the
	// same reason: a queued choice stops priority from passing, so the
	// state-check loop is exactly what does NOT run while one is open.
	g.pruneSacrificeChoicesLocked()
	g.pruneStaleZoneChangeChoicesLocked()
}

// removeObjectsOwnedByLocked drops every card owned by the departed
// player out of every zone the game tracks, and reports how many went.
//
// It removes them from the zone slices directly instead of routing
// them through MoveCard, and that is the whole point: leaving the game
// is not a zone change. No EventZoneMove, no EventLTB, no EventETB, so
// no dies trigger, no leaves-the-battlefield trigger, and no CR 603.10
// last-known-information snapshot for an event that never happened.
// The departing player's Blood Artist does not see their own board go.
//
// Every zone, not just the departed player's own four: a card they own
// can be on the shared battlefield under another player's control
// (stolen), in shared exile, on the stack, and — through a sandbox
// move or an effect that puts a card into another player's zone — in
// somebody else's hand or graveyard. Ownership is the test the rule
// names, so ownership is the test here.
//
// Caller must hold g.mu.
func (g *Game) removeObjectsOwnedByLocked(playerID uuid.UUID) int {
	removed := 0
	sweep := func(z *Zone) {
		if z == nil || len(z.Cards) == 0 {
			return
		}
		kept := make([]Card, 0, len(z.Cards))
		for _, c := range z.Cards {
			if c.Owner == playerID {
				removed++
				g.forgetObjectLocked(c.InstanceID)
				continue
			}
			kept = append(kept, c)
		}
		if len(kept) == len(z.Cards) {
			return
		}
		if len(kept) == 0 {
			z.Cards = nil
			return
		}
		z.Cards = kept
	}
	sweep(g.Battlefield)
	sweep(g.Stack)
	sweep(g.Exile)
	for _, p := range g.Seats {
		sweep(p.Library)
		sweep(p.Hand)
		sweep(p.Graveyard)
		sweep(p.Command)
	}
	if removed > 0 {
		g.recomputeSplitSecondLocked()
	}
	return removed
}

// forgetObjectLocked drops the per-object bookkeeping that outlives a
// card's zone — the maps keyed by instance ID. An ordinary zone change
// keeps all of it (a bounced planeswalker that comes back this turn
// has still activated a loyalty ability this turn); an object leaving
// the game keeps none of it, because there is no object left to
// remember anything about.
//
// Caller must hold g.mu.
func (g *Game) forgetObjectLocked(cardID uuid.UUID) {
	if g.StackMeta != nil {
		delete(g.StackMeta, cardID)
	}
	if g.LoyaltyActivatedThisTurn != nil {
		delete(g.LoyaltyActivatedThisTurn, cardID)
	}
	if g.lastKnownBattlefield != nil {
		delete(g.lastKnownBattlefield, cardID)
	}
	if g.lastKnownTriggerIdentity != nil {
		delete(g.lastKnownTriggerIdentity, cardID)
	}
}

// exileStillControlledLocked is CR 800.4a's last clause: once the
// departed player's own objects are gone and their control effects
// have ended, anything on the battlefield that is STILL controlled by
// them is exiled.
//
// What reaches this today is a permanent whose control baseline IS the
// departed player without a continuous effect saying so — another
// player's card that entered the battlefield under the departed
// player's control (reanimated out of an opponent's graveyard, put
// onto the battlefield off the top of their library). A Mind Control
// does NOT reach it: the Aura left with its owner one step ago, so the
// recompute has already handed the creature back (CR 613.1b), which is
// the printed behaviour and the reason the two steps are ordered.
//
// This one IS a zone change, so it goes through the ordinary
// battlefield exit and its triggers fire — the creature really is
// exiled from the battlefield, and a survivor's "whenever a creature
// leaves the battlefield" watcher saw it happen.
//
// Caller must hold g.mu.
func (g *Game) exileStillControlledLocked(playerID uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	var stuck []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == playerID {
			stuck = append(stuck, c.InstanceID)
		}
	}
	for _, id := range stuck {
		if err := g.executeBattlefieldLeaveLocked(id, ZoneExile, uuid.Nil, nil); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "leaving the game: exile of a still-controlled permanent failed: " + err.Error(),
			})
		}
	}
}

// exileGhostControlledLocked is CR 800.4c: a control-changing effect
// ending hands its permanent back to the player who controlled it
// before, and THAT player may have left the game in the meantime. The
// permanent is exiled rather than handed to a ghost.
//
// It is the same outcome CR 800.4a gives a permanent still controlled
// by the player who has just left, one effect-ending later — the
// difference is only when the effect ends. The A-steals-it, B-departs,
// A's-effect-ends order is what reaches this and not
// exileStillControlledLocked: at the moment B left, A controlled the
// permanent, so there was nothing to exile.
//
// The rules perform this as part of the effect ending; the engine
// performs it on the next state-based-action pass, which is the first
// moment after an effect ends that anything looks at the whole board.
// Reports how many it exiled, so the SBA pass can count itself as
// having fired and run again.
//
// A game that has ENDED keeps its final board (ADR 0060 Decision 5)
// and no state check runs on it, so this never strips one.
//
// Caller must hold g.mu.
func (g *Game) exileGhostControlledLocked() int {
	if g.Battlefield == nil || g.State != StateActive {
		return 0
	}
	var ghosts []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		p := g.playerByIDLocked(c.Controller)
		if p == nil || !p.Eliminated {
			continue
		}
		ghosts = append(ghosts, c.InstanceID)
	}
	for _, id := range ghosts {
		if err := g.executeBattlefieldLeaveLocked(id, ZoneExile, uuid.Nil, nil); err != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "CR 800.4c: exile of a permanent left to a departed controller failed: " + err.Error(),
			})
		}
	}
	return len(ghosts)
}

// dropDelayedTriggersForLocked removes the departed player's delayed
// triggered abilities (CR 603.7) from the queue. See the CR 800.4b/d
// note at the call site.
//
// Caller must hold g.mu.
func (g *Game) dropDelayedTriggersForLocked(playerID uuid.UUID) {
	if len(g.DelayedTriggers) == 0 {
		return
	}
	kept := g.DelayedTriggers[:0]
	for _, d := range g.DelayedTriggers {
		if d == nil || d.Controller != playerID {
			kept = append(kept, d)
		}
	}
	g.DelayedTriggers = kept
	if len(g.DelayedTriggers) == 0 {
		g.DelayedTriggers = nil
	}
}

// dropPendingTriggersForLocked removes every triggered ability waiting
// in the APNAP queue whose controller has left the game. Extracted
// from cleanupStackForEliminatedLocked so the departure can re-run it
// after the elimination event and the CR 800.4a exiles have had their
// chance to harvest one more.
//
// Caller must hold g.mu.
func (g *Game) dropPendingTriggersForLocked(playerID uuid.UUID) {
	if len(g.PendingTriggers) == 0 {
		return
	}
	kept := g.PendingTriggers[:0]
	for _, t := range g.PendingTriggers {
		if t == nil || t.Controller != playerID {
			kept = append(kept, t)
		}
	}
	g.PendingTriggers = kept
}
