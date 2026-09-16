package game

import (
	"github.com/google/uuid"
)

// attach.go is the S24 attachment relation (ADR 0036): the link an
// Equipment or an Aura holds to the permanent or player it is
// attached to, plus the two primitives that make and break it and
// the CR 704.5m/n state-based action that enforces it.
//
// The relation itself is one field — `Card.AttachedTo TargetRef` —
// stored on the ATTACHED object and pointing at the host. Everything
// in this file is a read or a write of that one field:
//
//	Attach   — AttachForEffect      (equip resolution, aura ETB, "attach")
//	Detach   — UnattachForEffect    (an effect that says "unattach")
//	Enforce  — attachmentSBALocked  (CR 704.5m/n, from the SBA loop)
//	Query    — AttachmentsOf / IsAttachedTo / AttachedHostOf
//
// Two rules make the shape work and both are worth stating up front:
//
//  1. Nothing sweeps the REVERSE direction. When a host leaves the
//     battlefield, the attachments pointing at it are left dangling
//     until the next state-based-action pass notices. That window is
//     load-bearing, not sloppy: Skullclamp's "whenever equipped
//     creature dies" trigger is harvested synchronously inside the
//     LTB emit, and it needs to still see itself attached.
//  2. Legality is checked against EFFECTIVE characteristics, never
//     printed ones — an Equipment attached to a creature that a
//     layer-4 effect stopped animating must fall off.

// IsAura reports whether the card is an Aura right now: an
// enchantment whose effective subtypes include "Aura". Post-layer on
// both halves, so an enchantment that became an Aura (or a
// permanent that became an enchantment) answers correctly.
//
// The distinction is not cosmetic. CR 704.5m puts an illegally
// attached Aura into its owner's graveyard; CR 704.5n merely
// unattaches an illegally attached Equipment and leaves it on the
// battlefield. Same trigger condition, opposite outcome — which is
// why the state-based action branches on exactly this predicate.
func (c Card) IsAura() bool {
	return c.IsEnchantment() && c.HasSubtype("Aura")
}

// IsAttached reports whether this card is currently attached to
// anything. The zero TargetRef (Kind == "") is the unattached
// sentinel, and TargetNone / TargetSelf are treated as unattached
// too — neither names a host.
func (c Card) IsAttached() bool {
	switch c.AttachedTo.Kind {
	case TargetCard, TargetPlayer:
		return c.AttachedTo.ID != uuid.Nil
	}
	return false
}

// IsAttachedTo reports whether this card is attached to the named
// card. The common shape in a catalog AppliesTo predicate —
// "equipped creature", "enchanted creature".
func (c Card) IsAttachedTo(hostID uuid.UUID) bool {
	return c.AttachedTo.Kind == TargetCard && hostID != uuid.Nil && c.AttachedTo.ID == hostID
}

// AttachForEffect attaches `attachmentID` to `host`, which may name
// a card (Equipment, most Auras) or a player (Curses). Both objects
// are looked up by ID rather than captured as pointers, per the
// standing rule for anything that runs off the stack.
//
// Re-attaching an already-attached permanent is legal and is what a
// second equip activation does (CR 702.6d): the old link is simply
// overwritten, and the CR 613.7d timestamp is refreshed.
//
// Attaching to a host that is not there, or to a card that is not on
// the battlefield, is refused rather than silently recorded — the
// state-based action would tear it down on the next pass anyway, and
// failing loudly at the primitive is easier to debug. CR 301.5c's
// "creature only" rule for Equipment is NOT enforced here: an effect
// that says "attach" to a non-creature is legal to perform, and the
// SBA unattaches it immediately afterwards, which is precisely how
// the rules sequence it.
//
// Caller must hold g.mu (this is a *ForEffect surface — it runs
// inside a resolution that already owns the write lock).
func (g *Game) AttachForEffect(attachmentID uuid.UUID, host TargetRef) error {
	if g.Battlefield == nil {
		return ErrCardNotFound
	}
	switch host.Kind {
	case TargetCard:
		if findCardOnBattlefield(g, host.ID) < 0 {
			return ErrCardNotFound
		}
	case TargetPlayer:
		if g.playerByIDLocked(host.ID) == nil {
			return ErrPlayerNotFound
		}
	default:
		return ErrInvalidParam
	}
	idx := findCardOnBattlefield(g, attachmentID)
	if idx < 0 {
		return ErrCardNotFound
	}
	c := &g.Battlefield.Cards[idx]
	if c.InstanceID == host.ID {
		// A permanent cannot be attached to itself (CR 301.5c).
		return ErrInvalidParam
	}
	c.AttachedTo = host
	c.AttachedAt = timeNowUnixNano()
	g.EmitEvent(Event{
		Kind:   EventAttach,
		Actor:  c.Controller,
		Source: c.InstanceID,
		CardID: c.InstanceID,
		Target: host.ID,
	})
	return nil
}

// UnattachForEffect breaks the link, leaving the permanent on the
// battlefield. Unattaching something that is not attached is a no-op
// rather than an error, so a catalog effect can call it
// unconditionally.
//
// This is the Equipment half of CR 704.5n by hand. It is NOT how an
// Aura leaves play — an Aura with no legal host goes to its owner's
// graveyard (CR 704.5m), which the state-based action handles.
//
// Caller must hold g.mu.
func (g *Game) UnattachForEffect(attachmentID uuid.UUID) error {
	idx := findCardOnBattlefield(g, attachmentID)
	if idx < 0 {
		return ErrCardNotFound
	}
	c := &g.Battlefield.Cards[idx]
	if !c.IsAttached() {
		return nil
	}
	host := c.AttachedTo
	c.AttachedTo = TargetRef{}
	c.AttachedAt = 0
	g.EmitEvent(Event{
		Kind:   EventUnattach,
		Actor:  c.Controller,
		Source: c.InstanceID,
		CardID: c.InstanceID,
		Target: host.ID,
	})
	return nil
}

// AttachedHostOf returns the host card of an attachment, or nil when
// it is unattached, attached to a player, or attached to something
// that is no longer on the battlefield. The pointer aliases into
// g.Battlefield.Cards and is only valid until the next mutation.
//
// Caller must hold g.mu (read or write).
func (g *Game) AttachedHostOf(c *Card) *Card {
	if c == nil || c.AttachedTo.Kind != TargetCard {
		return nil
	}
	idx := findCardOnBattlefield(g, c.AttachedTo.ID)
	if idx < 0 {
		return nil
	}
	return &g.Battlefield.Cards[idx]
}

// AttachmentsOf is the reverse lookup: every battlefield permanent
// currently attached to `hostID`. A linear scan of the battlefield
// rather than a second index, because the battlefield is tens of
// cards and a second data structure is a second thing that can
// disagree with the first.
//
// Caller must hold g.mu (read or write).
func (g *Game) AttachmentsOf(hostID uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil || hostID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].IsAttachedTo(hostID) {
			out = append(out, g.Battlefield.Cards[i].InstanceID)
		}
	}
	return out
}

// attachmentLegalLocked answers CR 704.5m/n's question for one card:
// may this attachment stay where it is?
//
// The two card kinds ask different questions, and flattening them
// would be a rules bug:
//
//   - An AURA's "enchant" clause is an ongoing legality condition
//     (CR 303.4c), so the answer is the card's own TargetSpec — the
//     literal same predicate the cast used at announce and again at
//     the CR 608.2b re-check. Evaluated WITHOUT the protection-style
//     keyword gate: CR 704.5m asks about the enchant restriction,
//     not about targeting, so a creature that gains hexproof does
//     not shrug off the Pacifism already on it.
//   - EQUIPMENT gets the flat CR 301.5c rule instead, and
//     deliberately not its equip clause: equip says "creature you
//     control", but that restriction applies only while activating.
//     Control of the creature may change afterwards and the
//     Equipment stays put; only "is it still a creature" matters.
//
// A card with no catalog TargetSpec (a sandbox Aura, a fixture)
// falls back to the same structural check as Equipment, which keeps
// the uncatalogued case sane rather than dropping it in a graveyard.
//
// Caller must hold g.mu. Must run AFTER the layer recompute — every
// type test here reads effective characteristics.
func (g *Game) attachmentLegalLocked(c *Card) bool {
	if c == nil {
		return true
	}
	if !c.IsAttached() {
		// CR 704.5m's OTHER half: "or is not attached to an object or
		// player". An Aura on the battlefield attached to NOTHING is
		// as illegal as one attached to something it may not enchant,
		// and the rule says so in the same sentence. Missing it left
		// a permanently-illegal board state reachable — an Aura put
		// onto the battlefield by an effect that does not say
		// "attached to" (Brilliant Restoration, Carmen) sat there
		// forever doing nothing, which no sequence of legal plays can
		// produce in paper.
		//
		// Scoped to Auras the catalog knows an enchant clause for,
		// for the same reason the attached branch below is: an
		// UNCATALOGUED Aura is a manual object in this sandbox. It is
		// cast with no target through the free-form picker
		// (attachResolvedAuraLocked documents that path), it attaches
		// to nothing, and a player is tracking it by hand. Sweeping
		// it into a graveyard would delete a card the table is using.
		// A catalogued Aura always acquires its host at resolution,
		// so reaching here means an effect put it onto the
		// battlefield without one.
		return !c.IsAura() || TargetSpecFor(CatalogKey(*c)) == nil
	}
	if c.IsAura() {
		if spec := TargetSpecFor(CatalogKey(*c)); spec != nil {
			return g.specMatchLocked(c.Controller, spec, c.AttachedTo, false)
		}
	}
	switch c.AttachedTo.Kind {
	case TargetPlayer:
		p := g.playerByIDLocked(c.AttachedTo.ID)
		return p != nil && !p.Eliminated
	case TargetCard:
		idx := findCardOnBattlefield(g, c.AttachedTo.ID)
		if idx < 0 {
			return false
		}
		// CR 301.5c — an Equipment can only be attached to a
		// creature. Effective, not printed: a creature that stopped
		// being a creature drops its sword.
		return g.Battlefield.Cards[idx].IsCreature()
	}
	return false
}

// attachmentSBALocked is the CR 704.5m / 704.5n state-based action.
// Returns true if it changed anything, which keeps the SBA loop
// spinning for another pass.
//
// Three outcomes, from two rules:
//
//	704.5n  Equipment attached to an illegal permanent  → unattach
//	704.5m  Aura attached to an illegal object/player   → graveyard
//	704.5m  Aura attached to NOTHING                    → graveyard
//
// Runs from stateBasedActionsLocked immediately after the layer
// recompute and before the destruction pre-pass, so that a creature
// which becomes lethally damaged BECAUSE its +2/+2 Aura fell off
// dies in the same settling rather than surviving a round.
//
// Illegal attachments are collected before any of them is acted on:
// routing an Aura to a graveyard mutates g.Battlefield.Cards
// underneath the iteration, and CR 704.3 wants the whole set
// evaluated against one snapshot of the game state anyway.
//
// Caller must hold g.mu.
func (g *Game) attachmentSBALocked() bool {
	if g.Battlefield == nil {
		return false
	}
	type doomedAttachment struct {
		id   uuid.UUID
		aura bool
	}
	var doomed []doomedAttachment
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		// NOT gated on IsAttached: CR 704.5m's condition is "attached
		// to an illegal object or player, OR not attached to an
		// object or player", and the second disjunct is the one an
		// unattached Aura fails. attachmentLegalLocked owns both.
		if g.attachmentLegalLocked(c) {
			continue
		}
		doomed = append(doomed, doomedAttachment{id: c.InstanceID, aura: c.IsAura()})
	}
	if len(doomed) == 0 {
		return false
	}
	for _, d := range doomed {
		// Re-find every pass: a previous iteration may have moved
		// cards out from under the index.
		idx := findCardOnBattlefield(g, d.id)
		if idx < 0 {
			continue
		}
		c := &g.Battlefield.Cards[idx]
		// An Aura that was never attached has no link to break and no
		// unattach to announce — emitting one would put a "became
		// unattached from nobody" line in the game log and bump the
		// layer version for a change that did not happen.
		if c.IsAttached() {
			host := c.AttachedTo
			controller := c.Controller
			c.AttachedTo = TargetRef{}
			c.AttachedAt = 0
			g.EmitEvent(Event{
				Kind:   EventUnattach,
				Actor:  controller,
				Source: d.id,
				CardID: d.id,
				Target: host.ID,
			})
		}
		if d.aura {
			// CR 704.5m. Routed through the normal battlefield-leave
			// path, so the CR 614 replacement pipeline and the
			// commander-zone built-in both apply — an enchantment
			// commander that is an Aura comes out right for free.
			_ = g.routeBattlefieldCardToOwnerGraveyardLocked(d.id)
		}
		// CR 704.5n is the else branch and it is already done: the
		// link is cleared and the Equipment stays put.
	}
	return true
}

// attachResolvedAuraLocked is the Aura half of "how attachment
// happens" (ADR 0036 decision 3a): an Aura is cast targeting
// (CR 303.4a) and enters the battlefield already attached to what it
// targeted. Called from resolveTopOfStackLocked once the permanent
// has landed, with the StackItem that carried the target.
//
// Keyed on the CARD TYPE, not on a catalog opt-in. Attachment is a
// rule of the type "Enchantment — Aura", not a property of an
// individual card, so an uncatalogued Aura played in the sandbox
// also attaches to whatever its controller picked. A `Spec.Enchant`
// opt-in would buy nothing the card's existing Spec.Targets clause
// does not already say.
//
// Silent on every path that does not apply — a non-Aura permanent,
// an Aura cast with no target (the free-form sandbox picker), an
// Aura with more than one target slot. The CR 608.2b re-check has
// already countered an Aura whose only target went away, so a
// surviving item's slot is known good.
//
// Caller must hold g.mu.
func (g *Game) attachResolvedAuraLocked(cardID uuid.UUID, item *StackItem) {
	if item == nil {
		return
	}
	idx := findCardOnBattlefield(g, cardID)
	if idx < 0 || !g.Battlefield.Cards[idx].IsAura() {
		return
	}
	var host TargetRef
	n := 0
	for _, t := range item.Targets {
		switch t.Kind {
		case TargetCard, TargetPlayer:
			host = t
			n++
		}
	}
	if n != 1 {
		return
	}
	_ = g.AttachForEffect(cardID, host)
}
