package game

// heal_damage.go — CR 701.69, heal, and the replacement built on it
// (#4614, Wolverine, Fierce Fighter).
//
// THE RULES.
//
//   - CR 701.69a: "To heal damage already dealt to a permanent, remove
//     that marked damage from that permanent. If an effect states that
//     damage already dealt to a permanent 'is healed,' that permanent's
//     controller removes all marked damage from that permanent."
//   - CR 614.1a: "If damage would be dealt to Wolverine, instead that
//     damage is dealt, but all other damage already dealt to him is
//     healed" is a replacement effect. It is not a prevention effect
//     (CR 615): the damage is dealt in full, so "damage can't be
//     prevented" leaves it alone and it changes no amount.
//   - CR 704.5g: lethal damage is judged on the damage marked on the
//     creature when state-based actions are checked. Because the heal
//     happens as the new damage is dealt, that is the new damage alone.
//   - CR 120.3d: damage from a source with wither or infect is -1/-1
//     counters, not marked damage, so it is neither "damage already
//     dealt" in the sense CR 701.69a heals nor healed by it. Counters
//     already placed stay.
//   - CR 702.2b / 704.5h: deathtouch destroys a creature that "has been
//     dealt damage by a source with deathtouch since the last time
//     state-based actions were checked". That is a fact about damage
//     dealt, not marked damage, so healing leaves it: the heal removes
//     the mark and Card.MarkedLethalByDeathtouch stays.
//
// SIMULTANEOUS DAMAGE. All combat damage of a step is dealt at the same
// time (CR 510.2), and one effect's damage to a creature (a fight, a
// "each creature" wipe) is one instruction; the damage "would be dealt"
// together. The engine opens one event per (source, recipient), applied
// one after another, so a replacement that healed on every event would
// let the second blocker's event heal the first's damage. "Damage already
// dealt" means dealt BEFORE this damage, so what the same DamageInstance
// (damage_instance.go) has already marked on the creature is not already
// dealt and is not healed. Card.damageInstance and damageInstanceMarked
// record what the latest instance marked on the card, written where the
// damage is marked (applyDamageToPermanentLocked) and cleared with it
// (clearBattlefieldDamage). Two instructions of one spell are two
// instances, so the second heals the first.
//
// Those two fields are not in a snapshot. A game restored in the middle of
// a combat damage step already gives the rest of that step a new instance
// (damage_instance.go, "What reaches a snapshot"), so the worst a restore
// can do is let a later event of the step heal damage the same step
// marked before the restore.
//
// If another replacement prevents or reduces the damage after this one is
// applied, the heal has already happened: this effect's own event is
// "that damage is dealt, and the other damage is healed", and the player
// who orders the effects (CR 616.1) can put the prevention first, in which
// case there is no damage to be dealt and this one does not apply.

// noteMarkedDamageInstance records that `n` damage from instance `inst`
// has just been marked on c. Instance 0 is "no instance" (an event built
// by hand) and is never matched.
func (c *Card) noteMarkedDamageInstance(inst DamageInstance, n int) {
	if n <= 0 {
		return
	}
	if inst != 0 && c.damageInstance == inst {
		c.damageInstanceMarked += n
		return
	}
	c.damageInstance, c.damageInstanceMarked = inst, n
}

// healableDamage is the marked damage on c that was dealt before the
// instance `inst`: all of it, less what `inst` itself has marked so far.
func (c *Card) healableDamage(inst DamageInstance) int {
	if c == nil || c.DamageMarked <= 0 {
		return 0
	}
	same := 0
	if inst != 0 && c.damageInstance == inst {
		same = min(c.damageInstanceMarked, c.DamageMarked)
	}
	return c.DamageMarked - same
}

// HealableDamageForEffect is how much marked damage on the damage event's
// recipient was dealt before the event's own damage (CR 701.69a, see the
// file comment): what HealOtherDamageForEffect would remove. Zero when
// the recipient is not on the battlefield or has none, which is how
// Wolverine's AppliesTo stays out of a CR 616 ordering prompt it has
// nothing to say in.
//
// Caller must hold g.mu.
func (g *Game) HealableDamageForEffect(ev *ReplacementEvent) int {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return 0
	}
	return findBattlefieldCard(g, ev.DamageTarget).healableDamage(ev.DamageInstance)
}

// HealOtherDamageForEffect is the Replace half of "if damage would be
// dealt to this creature, instead that damage is dealt, but all other
// damage already dealt to it is healed": it removes the damage marked on
// the event's recipient that was dealt before the event's own, leaving the
// event itself alone so the damage is dealt as it would have been.
// Returns how much was healed.
//
// Caller must hold g.mu (write).
func (g *Game) HealOtherDamageForEffect(ev *ReplacementEvent) int {
	if ev == nil || ev.Kind != RepEventDamage || ev.DamageAmount <= 0 {
		return 0
	}
	c := findBattlefieldCard(g, ev.DamageTarget)
	n := c.healableDamage(ev.DamageInstance)
	if n > 0 {
		c.DamageMarked -= n
	}
	return n
}
