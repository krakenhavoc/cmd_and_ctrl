package game

import (
	"slices"

	"github.com/google/uuid"
)

// excess_damage_tally.go — "if a creature or planeswalker an opponent
// controlled was dealt excess damage this turn" (Rith, Liberated
// Primeval, #5663).
//
// Excess damage (CR 120.10) is damage beyond lethal damage for a
// creature, and beyond its loyalty for a planeswalker. The engine lands
// one damage event at a time, so the fact is judged as each event lands,
// against the permanent as the damage finds it:
//
//   - a creature: the amount past what was still needed to make the
//     damage marked on it lethal (CR 120.6). Damage already marked
//     counts, so two blockers that together deal more than lethal damage
//     deal excess damage on the second event, whichever lands first, as
//     CR 120.10's "those sources together" asks. From a source with
//     deathtouch any amount greater than 1 is excess (CR 120.4a).
//   - a planeswalker: the amount past the loyalty it had before the
//     event. The first of two simultaneous events has already taken its
//     loyalty off, so the same sum falls out.
//
// A permanent that is both is dealt excess damage when either reading
// says so (CR 120.10, "the greatest of the calculated amounts"). A
// battle that is neither is not this tally's business.
//
// Combat or not, and whatever the source and whoever controls it: the
// printed condition asks only what the permanent was dealt. The amount
// judged is the damage dealt, after prevention and replacement.

// ExcessDamageRecord is one controller whose creature or planeswalker
// was dealt excess damage this turn. Controller is who controlled the
// permanent as it was dealt the damage ("an opponent controlled"), so
// the record outlives the permanent and a later change of control.
type ExcessDamageRecord struct {
	Controller uuid.UUID `json:"controller"`
	// Creature and Planeswalker say which reading found the excess; a
	// permanent that is both can set both.
	Creature     bool `json:"creature,omitempty"`
	Planeswalker bool `json:"planeswalker,omitempty"`
}

// creatureExcessDamage is how much of `amount` damage is excess for the
// creature c, read before the damage is marked. A creature whose
// toughness is not known has no lethal-damage threshold and is dealt no
// excess, which errs weaker.
func creatureExcessDamage(c *Card, amount int, deathtouch bool) int {
	if !c.ToughnessIsKnown() {
		return 0
	}
	need := max(c.CurrentToughness()-c.DamageMarked, 0)
	if deathtouch && need > 1 {
		need = 1
	}
	return max(amount-need, 0)
}

// planeswalkerExcessDamage is how much of `amount` damage is excess for
// the planeswalker c, read before the loyalty comes off.
func planeswalkerExcessDamage(c *Card, amount int) int {
	return max(amount-max(c.Counters[CounterLoyalty], 0), 0)
}

// recordExcessDamageLocked notes that `amount` damage, about to land on
// the battlefield permanent c, is excess damage to a creature or
// planeswalker, once per controller and reading. It must run before the
// damage is marked or the loyalty removed; the caller settles the
// layers first where a resolution may have left them stale
// (CurrentToughness reads the effective toughness). Caller must hold
// g.mu.
func (g *Game) recordExcessDamageLocked(c *Card, amount int, deathtouch bool) {
	rec := ExcessDamageRecord{
		Controller:   c.Controller,
		Creature:     c.IsCreature() && creatureExcessDamage(c, amount, deathtouch) > 0,
		Planeswalker: c.IsPlaneswalker() && planeswalkerExcessDamage(c, amount) > 0,
	}
	if !rec.Creature && !rec.Planeswalker {
		return
	}
	if slices.Contains(g.TurnTally.ExcessDamaged, rec) {
		return
	}
	g.TurnTally.ExcessDamaged = append(g.TurnTally.ExcessDamaged, rec)
}

// OpponentPermanentDealtExcessDamageThisTurn reports whether a creature
// or planeswalker that an opponent of `you` controlled was dealt excess
// damage this turn. Every other player is an opponent, one who has
// since left the game included. Caller must hold g.mu.
func (g *Game) OpponentPermanentDealtExcessDamageThisTurn(you uuid.UUID) bool {
	if you == uuid.Nil {
		return false
	}
	for _, r := range g.TurnTally.ExcessDamaged {
		if r.Controller != you {
			return true
		}
	}
	return false
}
