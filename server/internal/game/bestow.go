package game

import "github.com/google/uuid"

// bestow.go — CR 702.103, bestow (ADR 0141, #2862).
//
//	702.103a "Bestow [cost]" means "As you cast this spell, you may
//	choose to cast it bestowed. If you do, you pay [cost] rather than
//	its mana cost."
//	702.103b As a spell cast bestowed is put onto the stack, it becomes
//	an Aura enchantment and gains enchant creature. It is a bestowed
//	Aura spell, and the permanent it becomes as it resolves will be a
//	bestowed Aura. These effects last until the spell or the permanent
//	it becomes ceases to be bestowed.
//	702.103e As a bestowed Aura spell begins resolving, if its target is
//	illegal, it ceases to be bestowed and the effect making it an Aura
//	spell ends. It continues resolving as a creature spell.
//	702.103f If a bestowed Aura becomes unattached, it ceases to be
//	bestowed. If a bestowed Aura is attached to an illegal object or
//	player, it becomes unattached and ceases to be bestowed. This is an
//	exception to rule 704.5m.
//
// # The cast
//
// Bestow is an ordinary alternative cost (AlternativeCost.Bestow, built
// by effects.Bestow) whose Targets is the enchant creature clause. The
// announce path claims it like any other offer. Right after the claim,
// before every gate that reads the card, CastSpell sets Card.Bestowed on
// its working copy, and on the card on the stack once it is there. So
// the timing gate, the cast restrictions ("creature spells can't be
// cast") and the cast triggers all judge an Aura spell, which is CR
// 702.103d.
//
// # What a bestowed object is (CR 702.103b)
//
// An Aura enchantment that is not a creature. "Becomes an Aura
// enchantment" replaces the card types (CR 205.1a), so it loses
// Creature and the creature subtypes (CR 205.3d) and gains Enchantment
// and Aura. Every bestow card is already an enchantment creature, and
// supertypes are kept (Kestia stays legendary). The P/T and the
// abilities are the card's: a bestowed Nighthowler still prints "this
// creature and enchanted creature each get +X/+X", and the half that
// names the creature applies to nothing that is one.
//
// Off the battlefield (the spell on the stack) there is no layer pass,
// so the cold path of the type readers applies it to the printed
// baseline (bestowBaseline, read by printedShared, hasCardType,
// HasSubtype and HasAllCreatureTypes), as it does for a face-down
// object. On the battlefield it is a layer-4 effect from the board
// (bestowContinuousEffectsLocked), at the permanent's timestamp, so a
// later type-changing effect can still make the Aura a creature. The
// layer pass's own baseline (printedFresh) does not apply it, and
// neither do the copiable values: a Clone of a bestowed Aura copies an
// enchantment creature (CR 707.2).
//
// The enchant creature ability is not a keyword the engine lists. The
// target is the offer's clause, and the attachment's legality is read
// in attachmentLegalLocked: a bestowed Aura is legal attached to a
// creature on the battlefield.
//
// # Resolution (CR 702.103e, 608.3b)
//
// resolveTopOfStackLocked asks bestowTargetIllegalLocked before the CR
// 608.2b fizzle. A bestowed spell whose target is illegal is not
// countered: it ceases to be bestowed (the flag is cleared on the stack
// card and the targets dropped) and resolves on as a creature spell,
// entering the battlefield unattached.
//
// A bestowed spell that resolves normally enters with Bestowed seeded
// from the stack card (seedBestowedEntryLocked) before the Aura attach
// and before EventETB, so an enters trigger sees an Aura attached to its
// host and "whenever a creature enters" does not see a creature.
//
// # Unattached (CR 702.103f, 702.103g)
//
// When the host leaves or becomes illegal, the state-based action that
// would put an Aura into its owner's graveyard (CR 704.5m) instead
// unattaches it and clears Bestowed (attachmentSBALocked). It stays on
// the battlefield as an enchantment creature, and a 0/0 then dies to
// CR 704.5f in the same settling. Between the host leaving and that
// check, the layer-4 effect already treats it as unattached (CR 701.3d),
// as reconfigure does. UnattachForEffect clears it too. A bestowed Aura
// that phases in with no host is unattached and so is handled by the
// same check.

// BestowKey is the wire key of the bestow offer: CastSpellParams.
// AlternativeCost and StackItem.AltCost.
const BestowKey = "bestow"

// becomeBestowedAura is CR 702.103b on one characteristic: an Aura
// enchantment that is not a creature. Every slice it writes is a new
// one (loseCreatureType clips to zero capacity), so a cached baseline is
// never written through.
func (c *Characteristic) becomeBestowedAura() {
	c.loseCreatureType()
	if !typeListHas(c.Types, "enchantment") {
		c.Types = append(c.Types, "Enchantment")
	}
	if !typeListHas(c.Subtypes, "aura") {
		c.Subtypes = append(c.Subtypes, "Aura")
	}
}

// bestowBaseline is the cold-path characteristic of a bestowed object:
// the printed baseline with CR 702.103b applied. Only called for a card
// with Bestowed set and no layer cache.
func bestowBaseline(c *Card) Characteristic {
	in := printedInputsOf(c)
	ch := in.characteristic()
	ch.Controller = c.baseController()
	ch.becomeBestowedAura()
	return ch
}

// bestowedNowLocked reports whether CR 702.103b holds for a battlefield
// permanent right now: it is bestowed, and it is not attached to a
// permanent that has left the battlefield. A host that left unattached
// it (CR 701.3d), so it is a creature from that moment, before the
// state-based action clears the flag. An Aura that has just entered and
// is not yet attached is still bestowed.
//
// Caller must hold g.mu.
func (g *Game) bestowedNowLocked(c *Card) bool {
	if c == nil || !c.Bestowed {
		return false
	}
	if c.AttachedTo.Kind == TargetCard && c.AttachedTo.ID != uuid.Nil {
		return findCardOnBattlefield(g, c.AttachedTo.ID) >= 0
	}
	return true
}

// bestowedEffect is one bestowed Aura's CR 702.103b effect in layer 4.
type bestowedEffect struct {
	target    uuid.UUID
	timestamp int64
}

func (e bestowedEffect) Layer() (Layer, SubLayer) { return Layer4Type, 0 }
func (e bestowedEffect) Timestamp() int64         { return e.timestamp }
func (e bestowedEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.target
}
func (e bestowedEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	c.becomeBestowedAura()
}
func (e bestowedEffect) RemovesAbilities() bool      { return false }
func (e bestowedEffect) ContinuesAfterRemoval() bool { return false }

// bestowContinuousEffectsLocked is the layer pass's source list for CR
// 702.103b: one effect per bestowed permanent, at the time it entered
// (the effect began as the spell was put onto the stack, and the
// permanent's own timestamp is the earliest it can have here).
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) bestowContinuousEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !g.bestowedNowLocked(c) {
			continue
		}
		ts := c.EnteredBattlefieldAt
		if ts == 0 {
			ts = c.layerTimestamp()
		}
		out = append(out, bestowedEffect{target: c.InstanceID, timestamp: ts})
	}
	return out
}

// bestowTargetIllegalLocked is CR 702.103e's question as a bestowed
// spell begins to resolve: is it bestowed, and is its target illegal?
// The same re-check CR 608.2b runs (spellAllTargetsIllegalLocked), so
// a bestowed spell is never countered for its target and every other
// spell is judged exactly as before.
//
// Caller must hold g.mu.
func (g *Game) bestowTargetIllegalLocked(top *Card, item *StackItem) bool {
	if top == nil || !top.Bestowed || item == nil {
		return false
	}
	return spellAllTargetsIllegalLocked(g, item)
}

// endBestowOnStackLocked is CR 702.103e: the spell ceases to be
// bestowed and the effect making it an Aura spell ends. It continues
// resolving as a creature spell, so its targets go with it: CR 608.3a
// puts a permanent spell with no target onto the battlefield, and the
// Aura attach must not find one. Clears the flag on the stack card and
// on the resolver's copy.
//
// Caller must hold g.mu.
func (g *Game) endBestowOnStackLocked(top *Card, item *StackItem) {
	top.Bestowed = false
	if g.Stack != nil {
		for i := range g.Stack.Cards {
			if g.Stack.Cards[i].InstanceID == top.InstanceID {
				g.Stack.Cards[i].Bestowed = false
			}
		}
	}
	item.Targets = nil
}

// seedBestowedEntryLocked carries a bestowed spell onto the permanent
// it becomes (CR 702.103b: "the permanent it becomes as it resolves
// will be a bestowed Aura"). MoveCard clears the flag on every zone
// change, so the entry reads it off the stack card before the move and
// sets it here, after the CR 400.7 reset and before the Aura attach.
//
// Caller must hold g.mu.
func (g *Game) seedBestowedEntryLocked(entered uuid.UUID, moved *Card) {
	if c := findBattlefieldCard(g, entered); c != nil {
		c.Bestowed = true
		c.effective = nil
	}
	if moved != nil {
		moved.Bestowed = true
	}
	g.layerVersion.Add(1)
}

// bestowedAttachmentLegalLocked is attachmentLegalLocked for a bestowed
// Aura: enchant creature (CR 702.103b), so it is legal attached to a
// creature on the battlefield and illegal attached to anything else or
// to nothing. Protection has already been asked by the caller.
//
// Caller must hold g.mu. Effective characteristics.
func (g *Game) bestowedAttachmentLegalLocked(c *Card) bool {
	if c.AttachedTo.Kind != TargetCard {
		return false
	}
	host := findBattlefieldCard(g, c.AttachedTo.ID)
	return host != nil && host.IsCreature()
}

// unbestowLocked is CR 702.103f for one battlefield permanent: it
// ceases to be bestowed, and the layer pass makes it an enchantment
// creature again. The unattach is the caller's (attachmentSBALocked,
// UnattachForEffect), which emits EventUnattach.
//
// Caller must hold g.mu.
func (g *Game) unbestowLocked(c *Card) {
	if c == nil || !c.Bestowed {
		return
	}
	c.Bestowed = false
	c.effective = nil
	g.layerVersion.Add(1)
}
