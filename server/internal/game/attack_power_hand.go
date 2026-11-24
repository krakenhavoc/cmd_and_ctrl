package game

import (
	"strconv"

	"github.com/google/uuid"
)

// attack_power_hand.go is #6087's attack restriction: Ensnaring Bridge's
// "Creatures with power greater than the number of cards in your hand
// can't attack."
//
// CR 508.1c: "The active player checks each creature they control to see
// whether it's affected by any restrictions (effects that say a creature
// can't attack, or that it can't attack unless some condition is met). If
// any restrictions are being disobeyed, the declaration of attackers is
// illegal."
//
// It is an AttackTargetRestriction clause (PowerAtMostHandOf) rather than a
// CantAttack bit, for the reason OnlyAttackersMatching gives (#4601): the
// layer static that writes it runs inside the layer pass, where the
// creature's power is not final yet, and a bit set there could not see a
// card drawn or discarded after the pass. So the Bridge writes the clause
// onto EVERY creature, any controller's (its own controller's included),
// naming its controller, and the declaration reads both numbers live:
//
//   - the creature's power the comparison rules use (CR 208.1: layers and
//     counters, negative power included), and
//   - the number of cards in that player's hand right now.
//
// A creature with power equal to the hand size may attack ("greater than"
// is strict). The clause refuses every target, because it is a fact about
// the attacker, not about whom it attacks. Like every CR 508.1c check it is
// asked only at declaration: a creature already attacking keeps attacking
// when the Bridge controller's hand shrinks (CR 506.4a).
//
// The Bridge's controller is written by the layer pass, so a change of
// control is a new answer on the next pass; when the Bridge leaves the
// battlefield its static stops writing, and the restriction is gone.

// powerOverHandReasonLocked is the PowerAtMostHandOf clause: "" when it is
// unset or `attacker`'s power is at most the number of cards in that
// player's hand, else the reason in a player's words.
//
// Caller must hold g.mu with fresh layers.
func (r AttackTargetRestriction) powerOverHandReasonLocked(g *Game, attacker *Card) string {
	if r.PowerAtMostHandOf == uuid.Nil || attacker == nil {
		return ""
	}
	hand := 0
	if p := g.playerByIDLocked(r.PowerAtMostHandOf); p != nil && p.Hand != nil {
		hand = p.Hand.Size()
	}
	power := attacker.PowerForComparison()
	if power <= hand {
		return ""
	}
	return "its power (" + strconv.Itoa(power) + ") is greater than the number of cards in " +
		g.playerNameLocked(r.PowerAtMostHandOf) + "'s hand (" + strconv.Itoa(hand) + ")"
}
