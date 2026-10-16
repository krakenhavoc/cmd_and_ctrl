package game

import "github.com/google/uuid"

// stack_ability.go — activated abilities of a spell on the stack.
// ADR 0106 §1 amendment 2026-10-16, #1865.
//
// CR 113.6: "Abilities of an instant or sorcery spell usually function
// only while that object is on the stack." A catalog row says so the way
// a cycling row says it functions from a hand: ActivatedAbilityShape.Zones
// names ZoneStack (ADR 0062 Decision 1). Lightning Storm is the one
// printed card: "Discard a land card: Put two charge counters on
// Lightning Storm. You may choose a new target for it. Any player may
// activate this ability but only if Lightning Storm is on the stack."
//
// Nothing about the activation is new. The source is found in whatever
// zone holds it (findCardAndZoneLocked), the stack included; a spell has a
// controller (CR 112.2), so MayActivate answers for it as for a
// permanent; the costs, the timing and the item are the activation
// path's. What is the stack's own is what the EFFECT acts on: the spell,
// as long as it is still the object the ability was activated from
// (CR 400.7). AbilitySourceSpellForEffect is that read.

// HasStackAbility reports whether any activated ability `c` offers right
// now functions from the stack. The enumerator's and the view's fast
// negative for a spell: nearly none has one.
func HasStackAbility(c Card) bool {
	for _, ab := range ActivatedAbilitiesForCard(c) {
		if AbilityFunctionsFromZone(ab, ZoneStack) {
			return true
		}
	}
	return false
}

// AbilitySourceSpellForEffect is the spell an ability activated from the
// stack acts on: the item's source, while it is still on the stack as
// the SAME object the ability was activated from (CR 400.7). False when
// it has resolved, been countered, or left and come back (a spell put
// back on the stack by a later cast is a new object), and for an item
// whose source is not a spell. "Put two charge counters on Lightning
// Storm" then does nothing (CR 608.2b): there is no Lightning Storm to
// put them on, and counters on a card in a graveyard would be a counter
// on a different object.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilitySourceSpellForEffect(item *StackItem) (uuid.UUID, bool) {
	if item == nil || item.Kind == StackItemSpell || item.SourceCardID == uuid.Nil {
		return uuid.Nil, false
	}
	id := item.SourceCardID
	spell := g.StackMeta[id]
	if spell == nil || spell.Kind != StackItemSpell || g.Stack == nil {
		return uuid.Nil, false
	}
	for i := range g.Stack.Cards {
		c := &g.Stack.Cards[i]
		if c.InstanceID != id {
			continue
		}
		if ref := item.SourceObject; ref.ID == id && ref.Epoch != c.ObjectEpoch {
			return uuid.Nil, false
		}
		return id, true
	}
	return uuid.Nil, false
}
