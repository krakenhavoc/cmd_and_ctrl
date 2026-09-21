package game

// battlefield_entry.go — the ENTRY side's answer to the prune block
// the exit primitive runs (#1069).
//
// # The door that was not watched
//
// pruneCardSetChoicesLocked (#1045) withdraws a choose_cards prompt
// whose candidates have all left the zone it picks from, and trims the
// ones that have lost some. It ran at three sites and all three are
// EXITS or a departure: executeZoneRouteLocked (every routed move into
// a non-battlefield zone), the battlefield-leave resume
// (executeBattlefieldLeaveLocked) and the CR 800.4a sweep
// (leave_game.go).
//
// A card that leaves a hand, a library or a graveyard FOR THE
// BATTLEFIELD goes through none of them. routeDestinationLocked
// refuses a battlefield or stack destination outright — "not an exit,
// the entry path owns these" — so a graveyard pick whose candidate is
// reanimated under the open prompt, or a hand pick whose candidate is
// put onto the battlefield by a Warp World-style effect, kept offering
// a card its own resolver would refuse. Empty that list and the prompt
// is the #544 wedge the prune exists to close, arriving through the
// one door it did not watch.
//
// # One funnel, not three sprinklings
//
// The entry side has three landings and no shared finisher:
//
//	executeEntryToBattlefieldLocked   every RESUMABLE entry — the land
//	                                  play, stack resolution, a search,
//	                                  an exile return, a reanimation, a
//	                                  token (entry_choice.go)
//	putOntoBattlefieldFromZoneLocked  the hand / library "put onto the
//	                                  battlefield" batch and manifest
//	                                  (battlefield_put.go)
//	moveCardByRefLocked               the sandbox / admin move into the
//	                                  battlefield or the stack
//	                                  (mutations.go)
//
// The last two are the two sites that are deliberately NOT resumable
// (ReplacementEvent.entryResumable), which is exactly why they cannot
// be folded into the first. So the prune gets one named home that all
// three call, in the shape battlefieldExitLocked already has on the
// other side: a card added later is covered by the rule rather than by
// a code review, and a second prune that belongs to the entry side has
// somewhere to go.
//
// # What this is NOT
//
// It is not the exit block's three calls mirrored. Only the
// choose-cards prune has a question here:
//
//   - pruneSacrificeChoicesLocked asks whether a candidate is still on
//     the battlefield under its chooser's control. An entry only ADDS
//     permanents, so it can invalidate no option on an open sacrifice
//     prompt.
//   - pruneStaleZoneChangeChoicesLocked asks about a queued move of a
//     card OUT of the zone it has now left, and an entry does empty a
//     hand or a graveyard slot. It is not called here because #1069
//     scoped one prune at one door; the gap is recorded in
//     [ADR 0018](../../../docs/decisions/0018-triggers-on-the-stack.md)
//     §6's 2026-09-21 amendment rather than swept in blind, which is
//     the posture #1045 took towards this one.

// pruneChoicesAfterArrivalLocked runs the choose-cards prune for a card
// that has just ARRIVED somewhere the exit primitive does not route to
// — the battlefield, or the stack from the sandbox move.
//
// Call it AFTER the arrival is announced, last thing in the landing,
// for executeZoneRouteLocked's ordering reason: withdrawing a prompt
// runs the departure table's action for its kind, a run leg settling
// runs the rest of the printed instruction, and that continuation can
// queue the next prompt or start the next move. Everything this entry
// owes has to be finished before it does.
//
// It takes no card ID on purpose. The prune is keyed by ZONE and
// re-reads every open pick against the live board (#1045), so one call
// answers for a whole batch of arrivals, which is why
// putOntoBattlefieldFromZoneLocked calls it once after phase 3 rather
// than once per card.
//
// Caller must hold g.mu in write mode.
func (g *Game) pruneChoicesAfterArrivalLocked() {
	g.pruneCardSetChoicesLocked()
}
