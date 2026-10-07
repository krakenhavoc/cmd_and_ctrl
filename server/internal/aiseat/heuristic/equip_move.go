package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// equip_move.go — #2449: an equip that only moves an Equipment between
// the bot's own creatures buys what the new host gains, and no more.
//
// The flat activation price (ActivateBase plus OwnPermanentTarget for
// the creature it points at) presumes the ability does something for
// the bot. The first equip of an Equipment does. Moving it again does
// not, of itself: the Equipment leaves one of the bot's creatures to
// join another, and the board gains only the difference between the
// two hosts. Priced flat, a free equip (Lightning Greaves' Equip {0})
// beat passing from either side, so the bot moved it back and forth
// for ever, the CR 732 breaker named it, and the runner held on it
// with the table stopped (#810's hold).
//
// The rule: a move between two of the bot's own creatures is worth
// taking only when the new host, as it stands, is worth more than the
// current host is WITH the Equipment. Re-equipping the creature it is
// already on does nothing at all (CR 701.3b).
//
// Comparing the new host without the Equipment against the old host
// with it is what makes the rule settle. The wire carries each
// creature's current power, toughness and keywords, so the old host's
// value includes whatever the Equipment gives it and the new host's
// does not. A move needs v(new) > v(old with it), and the move back
// would need v(old) > v(new with it). For any Equipment that does not
// make its host worth less, both cannot hold, and along any chain of
// moves the hosts' values strictly rise, so the Equipment comes to
// rest. That is the guarantee #2449 needs; picking the best host is
// ADR 0126's "Equipment's best host", which stays out of scope. The
// haste case the issue names (moving Greaves onto a summoning-sick
// creature before combat) is not modelled: the bot keeps the Equipment
// where it is unless the new host is plainly the better creature.
//
// The first equip of an unattached Equipment, and moving one off a
// creature the bot no longer controls, are not lateral moves and keep
// the flat price.

// idleEquipMove prices an equip that gains nothing: below passing,
// which scores 0, like zeroXActivation and for the same reason: a
// repeatable move that buys nothing is a loop, and the CR 732 breaker
// stops a table that feeds one (ADR 0055).
const idleEquipMove = -1.0

// idleEquip reports whether this activation of src is an equip that
// moves src between two of the bot's own creatures, or onto the one it
// is already on, without the new host being the better creature.
func (st *state) idleEquip(src *protocol.CardView, cp activateParams) bool {
	row := rowAt(src, cp.AbilityIndex)
	if row == nil || !row.Equip {
		return false
	}
	host := st.attach.hostOf(src)
	if host == nil || host.Controller != st.me || !isCreature(host) {
		return false
	}
	var target *protocol.CardView
	for _, t := range cp.Targets {
		if t.Kind == "player" {
			continue
		}
		target = st.bf[t.ID]
		break
	}
	if target == nil {
		return false
	}
	if target.InstanceID == host.InstanceID {
		// CR 701.3b: attaching an Equipment to the object it is
		// already attached to does nothing.
		return true
	}
	if target.Controller != st.me {
		return false
	}
	return st.w.CreatureValue(target) <= st.w.CreatureValue(host)
}
