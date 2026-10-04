package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// excess_damage.go — "is dealt excess noncombat damage" (CR 120.10),
// shared by Magmatic Galleon (a Treasure) and Fall of Cair Andros
// (amass Orcs X, X the excess).
//
// "Excess" is not a field the damage event carries, so it is rebuilt
// from the creature: the harvester sees the event after the damage is
// marked (applyDamageToPermanentLocked runs before the event is
// emitted), so DamageMarked already includes this event's Amount, and
// subtracting the Amount back out recovers the damage marked BEFORE it.
// The excess is how far past lethal this event pushed the total, which
// is not merely "the creature is over lethal now": a creature that
// already had lethal damage marked and takes one more point has been
// dealt an excess of that one point, but an indestructible creature
// that was already over on an earlier, unrelated hit is not given
// credit for the old excess again.
//
// Read at the moment the trigger is harvested, which is when the
// numbers are still those of this event; Fall of Cair Andros carries
// the result on the stack item rather than rereading the board later.

// excessNoncombatDamageToOpponentCreature returns the excess noncombat
// damage that ev dealt to a creature `source`'s controller's opponents
// control, or zero when ev is not that.
func excessNoncombatDamageToOpponentCreature(ev game.Event, source *game.Card, g *game.Game) int {
	if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 {
		return 0
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !target.IsCreature() || target.Controller == source.Controller {
		return 0
	}
	toughness := target.CurrentToughness()
	after := target.DamageMarked - toughness
	if after <= 0 {
		return 0
	}
	before := (target.DamageMarked - ev.Amount) - toughness
	if before < 0 {
		before = 0
	}
	return after - before
}
