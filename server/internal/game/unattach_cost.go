package game

import (
	"slices"

	"github.com/google/uuid"
)

// unattach_cost.go — #4799, ADR 0020's 2026-11-02 amendment: "Unattach
// an Equipment from <this>" as part of an activated ability's cost.
//
//	Captain America, First Avenger  "Throw ... — {3}, Unattach an
//	                                Equipment from Captain America: He
//	                                deals damage equal to that
//	                                Equipment's mana value divided as you
//	                                choose among one, two, or three
//	                                targets."
//
// CR 701.3d: "To 'unattach' an Equipment from a creature means to move it
// away from that creature so the Equipment is on the battlefield but is
// not equipping anything." The payment is UnattachForEffect, the one
// primitive every unattach goes through, so EventUnattach fires and a
// "becomes unattached" watcher sees it.
//
// What the component shares with the other picks (ReturnToHand,
// ExilePermanents): the activator names the Equipment at announce
// (ActivateAbilityParams.UnattachIDs), everything is validated before
// anything is paid (ADR 0020 §3), and naming a permanent to pay a cost
// does not target it (CR 601.2h / 602.2b), so a hexproof Equipment pays.
//
// What it does not share:
//
//   - The candidates are the Equipment attached to the SOURCE, whoever
//     controls them (CR 301.5d: an Equipment's controller need not be the
//     equipped creature's). The printed words say "from Captain America",
//     not "you control".
//   - Nothing moves, so the pick is excluded from no other component.
//   - A divided amount may read it (DivideAmountArgs.Unattached): the
//     division is announced with this pick, in the one indivisible
//     announcement (CR 602.2b), so the amount is known when the gate
//     checks the shares.
//
// THE COMPONENT IS DATA (an int), so no func-typed route from Game is
// added (ADR 0041's closure ratchet).

// UnattachCostOptionsLocked lists the Equipment attached to `sourceID`
// that could pay an UnattachEquipment cost right now, in battlefield
// order: on the battlefield, with the subtype Equipment through the
// layered view (an Equipment that lost the subtype is not one, CR
// 301.5c), and attached to the source. The one candidate walk the
// validator, the protocol view and the legal-move enumerator read.
//
// Caller must hold g.mu (read or write).
func (g *Game) UnattachCostOptionsLocked(sourceID uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil || sourceID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == sourceID || !c.IsAttachedTo(sourceID) || !c.HasSubtype("Equipment") {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// UnattachCostOptionsForEffect is UnattachCostOptionsLocked on the
// *ForEffect surface, for the protocol view and the move enumerator.
func (g *Game) UnattachCostOptionsForEffect(sourceID uuid.UUID) []uuid.UUID {
	return g.UnattachCostOptionsLocked(sourceID)
}

// validateUnattachCostLocked checks the Equipment an activation names
// for its UnattachEquipment component, without unattaching anything.
//
//   - IDs sent for an ability with no such component are refused rather
//     than ignored, as a stray return_ids is.
//   - The source must be a permanent (effects.Register already refuses
//     the component off the battlefield; this is the activation's half).
//   - Exactly n Equipment, none named twice: a cost is not partially
//     payable (CR 118.3), so no Equipment attached means no activation.
//   - Each an Equipment attached to the source right now.
//
// Caller must hold g.mu.
func (g *Game) validateUnattachCostLocked(sourceID uuid.UUID, srcZone ZoneKind, n int, ids []uuid.UUID) error {
	if n <= 0 {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if srcZone != ZoneBattlefield || len(ids) != n {
		return ErrInvalidParam
	}
	options := g.UnattachCostOptionsLocked(sourceID)
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		if !slices.Contains(options, id) {
			return ErrIllegalTarget
		}
	}
	return nil
}

// payUnattachCostLocked unattaches the validated Equipment, one
// UnattachForEffect each (CR 701.3d). The Equipment stay on the
// battlefield under their controllers. Nil-safe: an ability without the
// component names none.
//
// Caller must hold g.mu.
func (g *Game) payUnattachCostLocked(ids []uuid.UUID) {
	for _, id := range ids {
		_ = g.UnattachForEffect(id)
	}
}
