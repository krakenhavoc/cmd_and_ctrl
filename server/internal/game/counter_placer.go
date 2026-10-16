package game

import "github.com/google/uuid"

// counter_placer.go — every counter placement names the player who puts
// the counters (#3942).
//
// # Why it matters
//
// "If YOU would put one or more counters" (Vorinclex, Monstrous Raider,
// Innkeeper's Talent, Lae'zel) and "whenever YOU put one or more
// counters" (Nest of Scarabs, Exemplar of Light, Terrasymbiosis) ask who
// is putting them, which is not who controls the permanent getting them:
// your spell putting a -1/-1 counter on an opponent's creature is a
// placement by you. ReplacementEvent.CounterPlacer and the emitted
// EventCounterPlaced.Actor carry the answer, but until this file only a
// handful of callers named one (CR 120.3d damage results, proliferate,
// counter costs, blight), and the readers guessed for the rest from the
// resolving slot and the event log, both of which outlive the
// resolution they describe.
//
// # The rule
//
// A placement that does not name its placer gets one here, before the
// CR 614 window opens, from the first of these that applies:
//
//  1. The innermost counterPlacerScope entry: an action that is not a
//     resolution names the player taking it. A mana ability's
//     activator (CR 605.3b: it resolves as it is activated, so nothing
//     is on the stack to ask) and the player taking a special action
//     (CR 116.2).
//  2. The controller of the spell or ability whose resolution is open
//     (CR 608.2c: its controller follows its instructions), including
//     its resolution-time prompts. The resolution is open from the moment the item begins
//     to resolve until the CR 704.3 boundary after it
//     (resolution_pause.go), which is the lifetime the issue asked for:
//     the resolving slot is cleared only at the next event batch, but
//     resolutionOpen closes at the boundary, so a mana ability or a
//     cost paid after the resolution is never credited to it. A
//     resolution paused on a prompt is restored from a snapshot without
//     the resolving slot, so the parked tail's item answers then.
//  3. The permanent's controller, or the player getting the counters.
//     That is CR 122.6a's answer for counters a permanent enters with,
//     CR 606.4's for a loyalty cost and CR 714.3c's for a Saga's lore
//     counter. Each of those names its placer explicitly anyway; this
//     arm is the backstop, so EventCounterPlaced.Actor is set for every
//     placement.
//
// A caller that knows better names the placer: "each player puts a
// counter on a creature they control" is each of those players, and an
// effect that says which player puts the counters a permanent enters
// with is that player (CR 122.6a). AddCounterByForEffect and its
// siblings take it.
//
// A removal names nobody. CR 614 replaces placements, not removals, and
// no reader asks who removed a counter.

// pushCounterPlacerLocked names `player` as the one putting any counters
// placed until the returned func is called, which the caller defers.
//
// Caller must hold g.mu.
func (g *Game) pushCounterPlacerLocked(player uuid.UUID) func() {
	g.counterPlacerScope = append(g.counterPlacerScope, player)
	n := len(g.counterPlacerScope)
	return func() {
		if len(g.counterPlacerScope) >= n {
			g.counterPlacerScope = g.counterPlacerScope[:n-1]
		}
	}
}

// openResolutionControllerLocked is the controller of the spell or
// ability whose resolution is open, or uuid.Nil when none is.
//
// Caller must hold g.mu.
func (g *Game) openResolutionControllerLocked() uuid.UUID {
	if !g.resolutionOpen {
		return uuid.Nil
	}
	if r := g.resolving; r != nil && r.item != nil {
		return r.item.Controller
	}
	if t := g.pausedResolutionTail; t != nil && t.item != nil {
		return t.item.Controller
	}
	return uuid.Nil
}

// counterPlacerForLocked is who puts counters on the card `cardID` or
// the player `playerID` (exactly one is set) when the placement did not
// say: the rule at the top of this file.
//
// Caller must hold g.mu.
func (g *Game) counterPlacerForLocked(cardID, playerID uuid.UUID) uuid.UUID {
	if n := len(g.counterPlacerScope); n > 0 {
		return g.counterPlacerScope[n-1]
	}
	if c := g.openResolutionControllerLocked(); c != uuid.Nil {
		return c
	}
	if playerID != uuid.Nil {
		return playerID
	}
	if c, ok := g.LookupCardForEffect(cardID); ok {
		return c.Controller
	}
	return uuid.Nil
}

// settleCounterPlacerLocked fills in the placer of a counter placement
// that did not name one. A no-op for any other event, for a removal and
// for a placement that already names its placer.
//
// Caller must hold g.mu.
func (g *Game) settleCounterPlacerLocked(ev *ReplacementEvent) {
	if ev == nil || ev.Kind != RepEventCounter || ev.CounterDelta <= 0 || ev.CounterPlacer != uuid.Nil {
		return
	}
	ev.CounterPlacer = g.counterPlacerForLocked(ev.CounterTarget, ev.CounterPlayer)
}
