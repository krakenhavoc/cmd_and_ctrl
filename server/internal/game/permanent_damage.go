package game

import "github.com/google/uuid"

// permanent_damage.go is the single answer to "what does damage DO to
// this permanent?" (CR 120.3), and it exists because before S27 there
// was only one answer and it was wrong for two of the three cases.
//
// Issue #406: every damage path in the engine incremented
// Card.DamageMarked, and the CR 704.5g lethal-damage state-based
// action compares DamageMarked to toughness INSIDE the
// `c.IsCreature()` branch. A planeswalker therefore accumulated a
// number nothing ever read: planeswalkers were unkillable by damage,
// silently, and a Lightning Bolt aimed at one was a two-mana no-op
// that looked like it had worked.
//
// CR 120.3 splits by what the permanent IS, and a permanent can be
// more than one thing at once:
//
//	120.3e  damage to a CREATURE is marked on it
//	120.3c  damage to a PLANESWALKER removes that many loyalty counters
//	120.3h  damage to a BATTLE removes that many defense counters
//
// So the clauses are additive, not a switch. A Gideon animated into a
// creature takes BOTH: the damage is marked on him and the loyalty
// comes off. Writing this as an if/else-if would have looked correct
// and quietly halved that case.
//
// TWO THINGS THIS DELIBERATELY DOES NOT DO.
//
// It does not route the counter removal through AddCounterForEffect.
// Removal is not addition: CR 614 replacement effects that watch
// counter placement — Doubling Season, Hardened Scales, Vorinclex —
// must not fire on damage taking loyalty OFF a walker, and routing
// through the effect helper would consult every one of them. The
// direct applyCounterLocked call with a negative delta still emits
// EventCounterPlaced, so anything watching the log still sees it.
//
// It does not redirect anything. Damage redirection from a player to
// a planeswalker they control was removed from the rules in 2018
// (the old CR 306.7); an attack on a planeswalker is declared against
// the planeswalker and lands there. There is nothing to redirect and
// no "choose a walker instead" prompt to build.

// applyDamageToPermanentLocked applies `amount` damage to the
// battlefield permanent `cardID` according to what it currently is
// (CR 120.3). Returns false when the card is not on the battlefield.
//
// The amount is assumed already post-replacement: every caller runs
// the CR 614 pipeline first and passes the settled number, which is
// what keeps Fog and the prevention shields in one place.
//
// `deathtouch` marks the CR 702.2c lethal flag on a creature. It is
// a parameter rather than a lookup because the non-combat caller
// (DealDamageToCreatureForEffect) has historically not applied it,
// and changing that is a separate question from this one.
//
// Caller must hold g.mu.
func (g *Game) applyDamageToPermanentLocked(cardID uuid.UUID, amount int, deathtouch bool) bool {
	if amount <= 0 {
		return false
	}
	idx := -1
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	c := &g.Battlefield.Cards[idx]
	isCreature := c.IsCreature()
	isPlaneswalker := c.IsPlaneswalker()
	isBattle := c.IsBattle()

	if isCreature {
		c.DamageMarked += amount
		if deathtouch {
			c.MarkedLethalByDeathtouch = true
		}
	}
	// The pointer must not survive a counter mutation: applyCounterLocked
	// emits an event, listeners run synchronously, and anything they do
	// can reallocate the battlefield slice.
	if isPlaneswalker {
		// CR 120.3c. A walker at 3 loyalty hit for 5 goes to 0, not to
		// -2: applyCounterLocked deletes the entry once it reaches
		// zero, and the CR 704.5i SBA sweeps it on the next check.
		_ = g.applyCounterLocked(cardID, CounterLoyalty, -amount)
	}
	if isBattle {
		// CR 120.3h, and the same clamping. The CR 704.5v SBA takes a
		// battle at zero defense; a Siege's defeated trigger fires off
		// that exit, not off this line.
		_ = g.applyCounterLocked(cardID, CounterDefense, -amount)
	}
	// A permanent that is none of the three — an ordinary artifact or
	// enchantment — takes the damage and nothing happens, which is the
	// rule (CR 120.3: damage to a permanent that is not a creature,
	// planeswalker or battle has no effect). Returning true anyway:
	// the damage WAS dealt, the event has already been emitted by the
	// caller, and lifelink still triggers off it.
	return true
}

// clearBattlefieldDamage wipes the damage marked on a permanent and
// the CR 702.2c deathtouch flag that rides with it. The counterpart of
// the marking above, and the only other thing that writes
// Card.DamageMarked outside a damage event.
//
// WHO IS ALLOWED TO CALL THIS, and it is a short list (#708). Damage
// marked on a permanent stays there until the cleanup step (CR 514.2);
// nothing else removes it except a permanent LEAVING the battlefield
// (where the damage belongs to an object that no longer exists) and a
// regeneration shield, which CR 701.19a says removes all damage from
// the permanent as part of the regeneration itself.
//
// So there are three callers (#816, #667). MoveCard's battlefield-exit cleanup
// is the one place a permanent physically leaves, whatever sent it —
// destroyed, sacrificed, exiled, bounced, tucked, milled or moved by
// hand — and CR 400.7 makes what lands in the new zone a new object,
// which must not arrive carrying the damage its previous existence
// took. Before #816 only the destroy path cleared it, so an exiled or
// bounced creature showed the number in its new zone and carried it
// back onto the battlefield when it was replayed. (The one entry that
// was safe is the exile → battlefield RETURN helper, which scrubs the
// card itself as it mints the new instance ID — a blink was fine and a
// recast was not, which is exactly the kind of per-path coverage this
// helper exists to end.) The other caller is the turn's CR 514.2 sweep
// (sweepTurnEndLocked, rotation.go), which clears every permanent at
// once and goes through here per card so "what clearing means" is
// written down once. The third is the CR 701.19 regeneration built-in
// (#667, regeneration.go), and it is the one caller that clears the
// damage while the permanent STAYS on the battlefield — it has to, or
// the CR 704.5g lethal-damage check destroys the creature again on the
// very next pass and spends every shield it has in a loop.
//
// It is deliberately NOT called from the destroy entry points: a
// destruction that a replacement effect rewrites into something else
// must leave the damage exactly where it was, both because the rule
// says so and because the replacement may want to read it. Nothing has
// moved while that window is open, so nothing is cleared — and that is
// why this is reached through the MOVE rather than through the
// destruction (#708).
//
// The deathtouch flag goes with the damage rather than being cleared
// on its own: CR 702.2c marks the creature as a consequence of damage
// dealt to it, so the two are one piece of per-turn state on one
// object.
func clearBattlefieldDamage(c *Card) {
	if c == nil {
		return
	}
	c.DamageMarked = 0
	c.MarkedLethalByDeathtouch = false
}
