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
	// Object is the permanent object that was dealt the excess damage, with
	// its CR 400.7 epoch, so "if that creature was dealt excess damage this
	// turn" (Maarika, Brutal Gladiator) can ask about one creature and not a
	// controller. The zero ref on a record written before it existed.
	// Additive within the snapshot schema.
	Object ObjectRef `json:"object,omitempty"`
}

// creatureExcessDamage is how much of `amount` damage is excess for the
// creature c, read before the damage is marked. A creature whose
// toughness is not known has no lethal-damage threshold and is dealt no
// excess, which errs weaker.
func creatureExcessDamage(c *Card, amount int, deathtouch bool) int {
	if !c.ToughnessIsKnown() {
		return 0
	}
	need := max(c.LethalDamageThreshold()-c.DamageMarked, 0)
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
//
// It returns the excess amount (CR 120.10: the greatest of the amounts
// calculated for a permanent that is both), zero when the damage is not
// excess; the damage event carries it as Event.ExcessDamage.
func (g *Game) recordExcessDamageLocked(c *Card, amount int, deathtouch bool) int {
	creatureExcess, walkerExcess := 0, 0
	if c.IsCreature() {
		creatureExcess = creatureExcessDamage(c, amount, deathtouch)
	}
	if c.IsPlaneswalker() {
		walkerExcess = planeswalkerExcessDamage(c, amount)
	}
	rec := ExcessDamageRecord{
		Controller:   c.Controller,
		Creature:     creatureExcess > 0,
		Planeswalker: walkerExcess > 0,
		Object:       ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch},
	}
	if !rec.Creature && !rec.Planeswalker {
		return 0
	}
	if !slices.Contains(g.TurnTally.ExcessDamaged, rec) {
		g.TurnTally.ExcessDamaged = append(g.TurnTally.ExcessDamaged, rec)
	}
	return max(creatureExcess, walkerExcess)
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

// ObjectDealtExcessDamageThisTurn reports whether the permanent object `ref`
// was dealt excess damage this turn (CR 120.10), and who controlled it as it
// was dealt that damage: "if that creature was dealt excess damage this turn"
// (Maarika, Brutal Gladiator). The fact outlives the permanent. A zero or
// unknown ref was dealt none. Caller must hold g.mu.
func (g *Game) ObjectDealtExcessDamageThisTurn(ref ObjectRef) (controller uuid.UUID, ok bool) {
	if ref.ID == uuid.Nil {
		return uuid.Nil, false
	}
	for _, r := range g.TurnTally.ExcessDamaged {
		if r.Object == ref {
			return r.Controller, true
		}
	}
	return uuid.Nil, false
}

// ObjectIDDealtExcessDamageThisTurn is ObjectDealtExcessDamageThisTurn for a
// caller that no longer holds the epoch: a creature that was dealt excess
// damage and has since died is a new object in the graveyard (CR 400.7), but
// the record still names the permanent it was. The latest record of the card
// wins. Caller must hold g.mu.
func (g *Game) ObjectIDDealtExcessDamageThisTurn(id uuid.UUID) (controller uuid.UUID, ok bool) {
	if id == uuid.Nil {
		return uuid.Nil, false
	}
	for i := len(g.TurnTally.ExcessDamaged) - 1; i >= 0; i-- {
		if r := g.TurnTally.ExcessDamaged[i]; r.Object.ID == id {
			return r.Controller, true
		}
	}
	return uuid.Nil, false
}
