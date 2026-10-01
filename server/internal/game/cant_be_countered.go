package game

import "github.com/google/uuid"

// cant_be_countered.go — every "can't be countered" in the engine, read
// at one gate. S23's "This spell can't be countered" rider (Supreme
// Verdict) started it; ADR 0106 §4 (#1806) gives the gate its full
// shape.
//
// NOT A KEYWORD, AND NOT A LAYER EFFECT. Every form of "can't be
// countered" is an effect that modifies the rules of the game rather
// than an object (CR 613.11), so it is applied after the layer pass
// and read at exactly one moment: when something tries to counter a
// spell. That is why none of it is stamped onto the stack item by the
// layer pass's stack step (ADR 0104), and why a static on a permanent
// reaches spells on the stack at all: it is not a characteristic of
// the spell, it is a statement the gate asks about the spell.
//
// THE FIVE SOURCES (ADR 0106 §4 decision 5), asked in this order, the
// first yes winning:
//
//  1. The spell's own printed rider — CatalogCantBeCountered.
//  2. The mana that paid for it (#1547): Cavern of Souls', Delighted
//     Halfling's and Boseiju's "that spell can't be countered" is a
//     spend rider stamped Applied on the item's payment record
//     (mana_spend_rider.go).
//  3. Marks on the spell's stack item — "target spell can't be
//     countered" (Vexing Shusher) and a spent one-use "the next spell
//     you cast" promise (Insist), StackItem.CantBeCountered
//     (counter_shield_grants.go).
//  4. The battlefield statics — "Spells you control can't be
//     countered" (Chimil, the Inner Sun) and its "you cast" and "any
//     player" forms, read off the battlefield on every ask
//     (counterShieldOnBattlefieldLocked, below).
//  5. A player's "this turn" grants on Player.Statics — Veil of
//     Summer, Bound // Determined, Domri, Anarch of Bolas's +1
//     (counterShieldGrantedLocked, counter_shield_grants.go). A
//     one-use promise on the same slice is NOT read here: it is spent
//     at the cast and becomes a source 3 mark.
//
// Every counter verb — CounterTargetForEffect, the to-zone and
// to-library counters, and the put_in_library prompt's
// counterableSpellOnStackLocked — asks spellCantBeCounteredLocked and
// nothing else, and the stack chip (#1553) is drawn from the same
// answer, so a new source added here reaches all of them at once.
//
// A spell that can't be countered is still a legal TARGET for a
// counterspell (CR 101.2: the "can't" only beats the counter itself).
// The counterspell resolves and does nothing; it does not fizzle.
//
// Not modelled, deliberately: the manual CounterSpell sandbox action
// honours none of these sources. It is the table's override, not a
// rules verb.

// CatalogCantBeCountered is the catalog hook the effects package
// wires at init, mirroring CatalogTargetSpec / CatalogModeSpec /
// CatalogAdditionalCost. Nil means no card declares the rider.
var CatalogCantBeCountered func(oracleID string) bool

// spellCantBeCounteredLocked reports whether the spell on the stack
// can't be countered, asking the sources in the order the file comment
// lists.
//
// Caller must hold g.mu with fresh layers (a battlefield static is
// keyed by CatalogAbilityKey and judged against its source's layered
// controller), as hexproofBypassedLocked's callers do. Reads only, so
// the view may ask under the read lock.
func (g *Game) spellCantBeCounteredLocked(spellID uuid.UUID) bool {
	item := g.StackMeta[spellID]
	spell := g.spellCardOnStackLocked(spellID)

	// 1. The spell's own printed rider.
	if spell != nil && spellPrintsCantBeCountered(*spell) {
		return true
	}
	// 2. The mana that paid for it (#1547).
	if item.SpellCantBeCounteredByMana() {
		return true
	}
	// 3. Marks on the item.
	if item != nil && len(item.CantBeCountered) > 0 {
		return true
	}
	// 4. The battlefield statics.
	if spell != nil && g.counterShieldOnBattlefieldLocked(item, *spell) {
		return true
	}
	// 5. The "this turn" grants on Player.Statics.
	return spell != nil && g.counterShieldGrantedLocked(item, *spell)
}

// spellCardOnStackLocked is the card a spell on the stack is, or nil.
//
// Caller must hold g.mu.
func (g *Game) spellCardOnStackLocked(spellID uuid.UUID) *Card {
	if g.Stack == nil {
		return nil
	}
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == spellID {
			return &g.Stack.Cards[i]
		}
	}
	return nil
}

// spellPrintsCantBeCountered reports whether the spell's own text says
// "This spell can't be countered". Keyed by CatalogKey, not
// CatalogAbilityKey: the rider is a statement on the spell, not an
// ability of a permanent (see CatalogAbilityKey's doc).
func spellPrintsCantBeCountered(spell Card) bool {
	if CatalogCantBeCountered == nil {
		return false
	}
	oracle := CatalogKey(spell)
	return oracle != "" && CatalogCantBeCountered(oracle)
}

// SpellCantBeCounteredForEffect is spellCantBeCounteredLocked for a caller
// that already holds the lock — the view projection (#1553), which shows
// the table why a Counterspell will do nothing. A stack item that is not
// a spell answers false: CR 101.2-style "can't be countered" is a
// statement about a spell, and an ability is countered by other verbs.
func (g *Game) SpellCantBeCounteredForEffect(id uuid.UUID) bool {
	item := g.StackMeta[id]
	if item == nil || item.Kind != StackItemSpell {
		return false
	}
	return g.spellCantBeCounteredLocked(id)
}

// --- source 4: the battlefield statics (ADR 0106 §4 decision 1) ----

// CounterShieldWhose is whose spells a "can't be countered" statement
// covers. The turn grants on Player.Statics take the same vocabulary,
// judged against the player the grant is on.
type CounterShieldWhose uint8

const (
	// CounterShieldYouControl is "Spells you CONTROL can't be
	// countered" (Chimil, the Inner Sun): a spell whose current
	// controller (StackItem.Controller, CR 112.2) is the controller of
	// the permanent with the static. A stolen spell (ADR 0104) is
	// covered by the thief's Chimil, not the caster's, and a copy you
	// control is a spell you control. The zero value, because it is
	// the narrower of the two "you" readings.
	CounterShieldYouControl CounterShieldWhose = iota

	// CounterShieldYouCast is "Spells you CAST … can't be countered"
	// (Thryx, the Sudden Storm; Cunning Nightbonder): a spell the
	// static's controller cast. The caster is StackItem.BaseController
	// (zero meaning Controller), which a change of control does not
	// move. A copy of a spell is not cast (CR 707.10), so it is never
	// covered.
	CounterShieldYouCast

	// CounterShieldAnyPlayer is the form with no "you" at all:
	// "Creature spells can't be countered" (Gaea's Herald), "Spells
	// can't be countered" (Lier, Disciple of the Drowned). Every
	// player's matching spell is covered, opponents' included.
	CounterShieldAnyPlayer
)

// CounterShieldStatic is one printed "<these> spells can't be countered"
// static on a permanent. Build it with the constructors in
// cards/effects/counter_shields.go rather than by hand.
//
// It is catalog data and is never stored: the gate reads it off the
// battlefield every time it asks (CR 611.3a, a static's effect "isn't
// locked in"), so it covers a spell cast before the permanent arrived,
// stops covering anything the moment the permanent leaves, and two
// of them compose for free.
type CounterShieldStatic struct {
	// Label is the clause as printed, for the log.
	Label string

	// Whose is whose spells the static covers.
	Whose CounterShieldWhose

	// Spell narrows which spells — "creature", "green", "with power 5
	// or greater", "Sliver", "with mana value 5 or greater". `spell` is
	// the card on the stack as it stands; `source` is the permanent
	// with the static, live off the battlefield. Nil covers every
	// spell. Treat the game as read-only.
	Spell func(g *Game, spell Card, source *Card) bool
}

// CatalogCounterShields returns the "spells can't be countered"
// statics a battlefield permanent with the given catalog key has.
// carddef.go sets it from CardDef.SpellsCantBeCountered; a
// game-package test may stub it directly.
var CatalogCounterShields func(key string) []CounterShieldStatic

// CounterShieldsForCard is the shields a permanent has right now: its
// catalog entry's, keyed by CatalogAbilityKey so a Chimil that has
// lost all its abilities (CR 613.1f) shields nothing. A phased-out
// permanent is not in the battlefield slice (phasing.go), so the gate
// never reaches one (CR 702.26b).
func CounterShieldsForCard(c Card) []CounterShieldStatic {
	if CatalogCounterShields == nil {
		return nil
	}
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	return CatalogCounterShields(key)
}

// counterShieldOnBattlefieldLocked reports whether some permanent's
// static says this spell can't be countered (CR 604.2: active while
// the permanent is on the battlefield and has the ability).
//
// `item` is the spell's stack item; nil (a spell with no item, which
// only a hand-built test makes) is judged as having no controller, so
// only an any-player shield can cover it.
//
// Caller must hold g.mu.
func (g *Game) counterShieldOnBattlefieldLocked(item *StackItem, spell Card) bool {
	if CatalogCounterShields == nil || g.Battlefield == nil {
		return false
	}
	var controller, caster uuid.UUID
	copied := false
	if item != nil {
		controller, caster = item.Controller, item.BaseController
		if caster == uuid.Nil {
			caster = controller
		}
		copied = item.IsCopy
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		for _, s := range CounterShieldsForCard(*src) {
			switch s.Whose {
			case CounterShieldYouControl:
				if controller == uuid.Nil || controller != src.Controller {
					continue
				}
			case CounterShieldYouCast:
				if copied || caster == uuid.Nil || caster != src.Controller {
					continue
				}
			case CounterShieldAnyPlayer:
			default:
				continue
			}
			if s.Spell == nil || s.Spell(g, spell, src) {
				return true
			}
		}
	}
	return false
}
