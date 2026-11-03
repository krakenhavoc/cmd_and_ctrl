package game

import "github.com/google/uuid"

// cant_have_counters.go is #4692: "Enchanted creature … can't have
// counters put on it" (Blossombind).
//
// A prohibition, not a replacement. CR 101.2: "can't" beats every
// instruction to put a counter, so the placement simply does not happen,
// and because there is no event there is nothing for a CR 614 counter
// replacement to see: Doubling Season, Hardened Scales and Vizier of
// Remedies never apply, no CR 616 ordering question is asked, and
// nothing triggers on "whenever one or more counters are put". Removing
// counters is untouched; the permanent keeps the ones it already has.
//
// The flag is Characteristic.CantHaveCounters, written by an ordinary
// layer-6 static of the restricting permanent (the Aura's, for
// Blossombind), so it ends the moment the Aura leaves and a "loses all
// abilities" on the creature does not lift it. It is read in these
// places, one predicate each, so nothing offers what the engine refuses
// (#544):
//
//   - Every placement door. counterPlacementRefusedLocked is asked
//     before the CR 614 window opens: AddCounterByKeyedForEffect and
//     the starting-loyalty stamp (placeCountersThroughReplacementsLocked),
//     every cost, Saga and mana-ability placement
//     (addCounterMustSettleNowLocked), and a damage result's -1/-1
//     counters from wither or infect (placeDamageResultCountersLocked,
//     CR 120.3d). Entering with counters (CR 122.6) and proliferate's
//     placements go through the first of these. A refused placement is
//     a terminal outcome told zero, so "move a counter" puts nothing
//     on and so removes nothing from the first object (CR 122.5), and
//     the resume of a placement that paused on a CR 616 prompt asks
//     again before it lands (applyResolvedCounterThenLocked).
//   - The counter door itself, applyCounterByLocked, which refuses a
//     placement that reached it any other way (a sandbox edit, a
//     face-down turn's counter).
//   - COSTS (CR 118.3). A cost that puts a counter on it can't be paid:
//     Devoted Druid's (canPlaceCounterLocked), a +N loyalty cost (the
//     activation's CR 606.6 check and the enumerator's), and a blight
//     cost can't choose it (BlightOptionsForEffect,
//     validateBlightCreatureLocked; CR 701.68b).
//   - PROLIFERATE's choices leave it out (ProliferateCandidatesForEffect),
//     since choosing it would give it nothing.
//
// The view ships it as `cant_have_counters` on the card.

// cantHaveCountersLocked reports whether `c`, a permanent on the
// battlefield, can't have counters put on it right now. False for nil
// and for a card in any other zone.
//
// Reads the effective characteristic, so the caller relies on fresh
// layers, as every other static gate does. Caller must hold g.mu.
func (g *Game) cantHaveCountersLocked(c *Card) bool {
	if c == nil || g.Battlefield == nil || !g.Battlefield.Contains(c.InstanceID) {
		return false
	}
	return c.Effective().CantHaveCounters
}

// counterPlacementRefusedLocked reports whether `ev` is a counter
// PLACEMENT (a positive delta on a permanent, not a player) onto a
// permanent that can't have counters put on it. Caller must hold g.mu.
func (g *Game) counterPlacementRefusedLocked(ev *ReplacementEvent) bool {
	if ev == nil || ev.Kind != RepEventCounter || ev.CounterPlayer != uuid.Nil || ev.CounterDelta <= 0 {
		return false
	}
	return g.cantHaveCountersLocked(findBattlefieldCard(g, ev.CounterTarget))
}

// CantHaveCountersForEffect is cantHaveCountersLocked by instance ID, for
// the view and the catalog. Caller must hold g.mu (read is enough).
func (g *Game) CantHaveCountersForEffect(cardID uuid.UUID) bool {
	return g.cantHaveCountersLocked(findBattlefieldCard(g, cardID))
}
