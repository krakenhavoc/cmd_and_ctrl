package game

import "github.com/google/uuid"

// battlefield_exit.go — the Game-side half of a permanent leaving the
// battlefield (CR 400.7).
//
// MoveCard (zone.go) is the card-side half: one battlefield-exit
// cleanup that strips the leaving CARD of everything that belonged to
// the permanent it stopped being — tapped, counters, marked damage,
// attachments, the controller baseline, the chosen tribe and colour.
// That block cannot reach state the ENGINE keeps about the object,
// because MoveCard is a package-level function over two zones with no
// Game in sight.
//
// So per-object state that lives in a map on Game — keyed by instance
// ID, which survives a zone change — needs the same clearing from a
// place that has the Game. This is that place, and it is one function
// for the same reason MoveCard's block is one block: every battlefield
// exit has to forget the same things, and a second copy is a second
// thing to forget to update.

// battlefieldExitLocked runs as a permanent leaves the battlefield,
// with the card still in g.Battlefield: it takes the CR 603.10
// last-known-information snapshot the leave-the-battlefield triggers
// are judged on, and forgets the per-object state the Game was
// keeping about the object that is ending.
//
// Called from each of the three battlefield exits — the destroy /
// sacrifice route (executeBattlefieldLeaveLocked), the general
// effect route (executeZoneRouteLocked) and the sandbox move
// (moveCardByRefLocked, behind MoveCardByID) — immediately before
// their MoveCard.
//
// Caller must hold g.mu.
func (g *Game) battlefieldExitLocked(cardID uuid.UUID) {
	g.snapshotLKILocked(cardID)
	g.forgetPerObjectTurnStateLocked(cardID)
}

// forgetPerObjectTurnStateLocked drops every "this object did this"
// entry the engine holds against cardID.
//
// CR 400.7: an object that moves to another zone becomes a NEW
// object, with no memory of the one that left. The engine's per-object
// registries are keyed by instance ID, and an instance ID survives a
// zone change — a planeswalker bounced and recast keeps its own — so
// nothing else takes these entries away until the turn ends.
//
// That was #630: activate Teferi's +1, return him to hand with Venser,
// recast him, and the server still refused the loyalty ability and the
// client still greyed both rows with "Already activated this turn".
// The permanent that came back is a new object and CR 606.3 applies
// per object, so it may activate one.
//
// What is forgotten here, and why each one is per OBJECT:
//
//   - LoyaltyActivatedThisTurn — CR 606.3's "only one loyalty ability
//     of a permanent, and only once each turn". The bug.
//   - announcedAttacks / announcedBlocks / announcedBecameBlocked —
//     what this combat has already announced about this creature
//     (#830, #859). clearCombatLocked drops them at end of combat, so
//     a stale entry can only be read by the same combat the permanent
//     left, which nothing in the engine can reach today: a permanent
//     that leaves is removed from combat and comes back with no
//     attacking or blocking target. They are here because they are the
//     same shape and the same rule, not because they are reachable.
//
// What deliberately stays:
//
//   - TurnTally's per-ability counts (Resolved, Triggered, LoopRun,
//     LoopAllowance) are keyed by TallyKey(source, label), which is
//     per object too, and a returning object arguably should trigger
//     its "only once each turn" ability again. Clearing them here
//     would hand ADR 0055's loop breaker an escape hatch: a blink loop
//     leaves and re-enters on every iteration, so it would reset its
//     own LoopRun every iteration and never trip the threshold. That
//     trade needs the breaker keyed on something other than the source
//     instance before it can be made; see turn_tally.go.
//   - oncePerBatchFired is per object, but each entry is compared
//     against the live event batch (#829), so an entry left by an
//     object that has gone cannot match a later batch.
//   - lastKnownBattlefield / lastKnownTriggerIdentity are the CR 603.10
//     snapshots this very exit writes, consumed by the harvester a few
//     lines later.
//
// Caller must hold g.mu.
func (g *Game) forgetPerObjectTurnStateLocked(cardID uuid.UUID) {
	delete(g.LoyaltyActivatedThisTurn, cardID)
	if len(g.LoyaltyActivatedThisTurn) == 0 {
		g.LoyaltyActivatedThisTurn = nil
	}
	delete(g.announcedAttacks, cardID)
	if len(g.announcedAttacks) == 0 {
		g.announcedAttacks = nil
	}
	delete(g.announcedBlocks, cardID)
	if len(g.announcedBlocks) == 0 {
		g.announcedBlocks = nil
	}
	delete(g.announcedBecameBlocked, cardID)
	if len(g.announcedBecameBlocked) == 0 {
		g.announcedBecameBlocked = nil
	}
}
