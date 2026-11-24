package game

import "github.com/google/uuid"

// cast_gate_x.go — the announced X reaches the cast gate (#6085).
//
// Lavinia, Azorius Renegade: "Each opponent can't cast noncreature
// spells with mana value greater than the number of lands that player
// controls." The spell's mana value the ban reads is its mana value on
// the stack, and there CR 202.3e counts {X} at the number chosen for it:
//
//	202.3e When calculating the mana value of an object with an {X} in
//	its mana cost, X is treated as 0 while the object is not on the
//	stack, and X is treated as the number chosen for it while the
//	object is on the stack.
//
// X is chosen as the spell is proposed (CR 601.2b), and CR 601.3a lets
// the caster begin the cast whenever a choice made while proposing it
// could take it out of the ban — so the gate judges the cast with the
// X it announces, and a Fireball with X too large for the caster's
// lands is refused at CastSpell (CR 601.2e), never cast.
//
// Before #6085 the gate saw X as 0, and an opponent could cast an X
// spell for any X: weaker than printed. CastQuery now carries XValue,
// and a restriction on mana value declares its CEILING as data
// (CastRestriction.ManaValueCeiling) rather than hiding the comparison
// in Forbids, for the reason SpellXCeilingLocked exists: the enumerator
// and the view ask the gate with X=0 ("is ANY cast open", which X=0,
// the least mana value, answers), and must then know how far X may go
// so the bot is never offered, and the client's X picker never stops
// at, an X the gate refuses (#544). One function, CastGateXCeilingLocked,
// answers that for all of them, from the same restrictions the gate
// reads.

// SpellManaValue is the mana value the spell this query proposes will
// have on the stack: its mana value with each {X} counted at the
// announced XValue (CR 202.3e). ok is false when the printed cost can't
// be read.
func (q CastQuery) SpellManaValue() (mv int, ok bool) {
	cost, err := ParseCost(manaCostForValue(q.Card))
	if err != nil {
		return 0, false
	}
	return cost.ManaValueWithX(q.XValue), true
}

// refusesManaValue reports whether the restriction's ManaValueCeiling
// refuses the cast `q` proposes. A spell whose cost can't be read is
// refused rather than read as mana value 0, which would pass any
// ceiling (the reason Card.ParsedManaValue reports ok).
func (r CastRestriction) refusesManaValue(q CastQuery) bool {
	if r.ManaValueCeiling == nil {
		return false
	}
	ceiling, ok := r.ManaValueCeiling(q)
	if !ok {
		return false
	}
	mv, ok := q.SpellManaValue()
	return !ok || mv > ceiling
}

// CastGateXCeilingLocked is the largest X `caster` may announce for
// `card` cast out of `zone` before a ManaValueCeiling on the board or
// on an emblem refuses it, read now. ok is false when no such ceiling
// bears on the cast, or when the card's mana cost has no {X} (X then
// cannot move its mana value). The answer is 0 when the ceiling is
// already below the spell's mana value at X=0; the gate itself refuses
// that cast at any X, so a caller asks CastGateLocked first.
//
// Restrictions are read exactly as CastGateLocked reads them (an
// ignored Ignorable restriction bears on nothing), so the number and
// the refusal cannot disagree.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) CastGateXCeilingLocked(caster uuid.UUID, card Card, zone ZoneKind) (ceiling int, ok bool) {
	if CatalogCastRestrictions == nil {
		return 0, false
	}
	cost, err := ParseCost(manaCostForValue(card))
	if err != nil || cost.XSlots == 0 {
		return 0, false
	}
	base := cost.ManaValue()
	q := CastQuery{Game: g, Card: card, Controller: caster, FromZone: zone}
	fold := func(src *Card) bool {
		for _, r := range castRestrictionsOf(src) {
			if r.ManaValueCeiling == nil {
				continue
			}
			if r.Ignorable && g.effectIgnoredLocked(caster, src) {
				continue
			}
			q.Source = *src
			n, applies := r.ManaValueCeiling(q)
			if !applies {
				continue
			}
			x := max(0, (n-base)/cost.XSlots)
			if !ok || x < ceiling {
				ceiling, ok = x, true
			}
		}
		return true
	}
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			fold(&g.Battlefield.Cards[i])
		}
	}
	g.forEachCommandSourceLocked(fold)
	return ceiling, ok
}

// CastGateXCeilingForEffect is CastGateXCeilingLocked on the *ForEffect
// surface, for the legal-move enumerator and the protocol view.
//
// Caller must hold g.mu (read or write) with fresh layers.
func (g *Game) CastGateXCeilingForEffect(caster uuid.UUID, card Card, zone ZoneKind) (int, bool) {
	return g.CastGateXCeilingLocked(caster, card, zone)
}
