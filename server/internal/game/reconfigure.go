package game

import "github.com/google/uuid"

// reconfigure.go — CR 702.151, reconfigure (#2639).
//
//	702.151a Reconfigure represents two activated abilities.
//	Reconfigure [cost] means "[Cost]: Attach this permanent to another
//	target creature you control. Activate only as a sorcery" and
//	"[Cost]: Unattach this permanent. Activate only if this permanent
//	is attached to a creature and only as a sorcery."
//	702.151b Attaching an Equipment with reconfigure to another
//	creature causes the Equipment to stop being a creature until it
//	becomes unattached from that creature.
//
// # The two abilities
//
// Both are ordinary catalog activated abilities, built by one
// constructor in the effects package (effects.Reconfigure) and marked
// with ActivatedAbilityShape.Reconfigure. The attach half resolves
// through AttachSourceForEffect, exactly as equip does. The unattach
// half resolves through UnattachSourceForEffect, below. Neither is an
// equip ability: Leonin Shikari's "equip abilities" does not reach
// them, which is why the bit is its own and not Equip.
//
// # "Stops being a creature" (CR 702.151b)
//
// A layer-4 type-changing effect, derived from the board rather than
// stored. One effect per battlefield permanent that prints reconfigure
// and is attached to a permanent that is still on the battlefield. It
// removes the creature card type and, with it, the creature subtypes
// (CR 205.3d: an object can't have a subtype that doesn't correspond
// to one of its types), so a Lizard Blades on a creature is an
// "Artifact — Equipment". Artifact and Equipment stay.
//
// Its timestamp is the attach (Card.AttachedAt), which is when the
// rule's effect began. A later layer-4 effect that makes artifacts
// creatures (March of the Machines) therefore makes it a creature
// again, and CR 301.5c's reconfigure exception keeps it attached.
//
// Derived rather than recorded on attach, which costs one honest
// difference from the rule's wording: the rule keys the effect on the
// Equipment having reconfigure WHEN it was attached, and this keys it
// on the card printing reconfigure. The two differ only for a
// permanent that gains or loses reconfigure while attached, and the
// derivation reads the PRINTED rows, never the layered ones, so an
// attached Equipment that loses all abilities stays a noncreature, as
// the rule says it does. No card grants reconfigure. Nothing is added
// to the snapshot.
//
// The host is checked too. CR 701.3d says an Equipment whose host left
// the battlefield has become unattached, even while the link waits
// for the state-based action to clear it, so the Equipment is a
// creature again from that moment.
//
// # CR 301.5c
//
// "An Equipment that's also a creature can't equip a creature unless
// that Equipment has reconfigure." attachmentLegalLocked asks
// equipmentCreatureMayEquipLocked, so an animated Equipment without
// reconfigure becomes unattached as a state-based action (CR 704.5n),
// and one with reconfigure stays.

// printsReconfigure reports whether the card's own catalog entry
// declares a reconfigure ability. The printed rows, not the layered
// ones: see the file comment for why CR 702.151b reads these.
func printsReconfigure(c *Card) bool {
	if c == nil || CatalogActivatedAbilities == nil {
		return false
	}
	key := catalogKeyOf(c)
	if key == "" {
		return false
	}
	for _, ab := range CatalogActivatedAbilities(key) {
		if ab.Reconfigure {
			return true
		}
	}
	return false
}

// HasReconfigure reports whether the permanent has a reconfigure
// ability right now: one of the rows ActivatedAbilitiesForCard offers
// is marked Reconfigure. A permanent that has lost all its abilities
// has none. This is CR 301.5c's "has reconfigure".
//
// Caller must hold g.mu (read or write); the layer pass must have run.
func HasReconfigure(c *Card) bool {
	if c == nil {
		return false
	}
	for _, ab := range activatedAbilitiesOf(c) {
		if ab.Reconfigure {
			return true
		}
	}
	return false
}

// equipmentCreatureMayEquipLocked is CR 301.5c's reconfigure clause for
// one attached Equipment: false when it is a creature right now and has
// no reconfigure ability.
//
// Caller must hold g.mu. Effective characteristics.
func equipmentCreatureMayEquipLocked(c *Card) bool {
	if !c.IsCreature() {
		return true
	}
	return HasReconfigure(c)
}

// reconfiguredAttachedLocked reports whether CR 702.151b holds for c
// right now: c prints reconfigure and is attached to a permanent that
// is still on the battlefield.
//
// Caller must hold g.mu.
func (g *Game) reconfiguredAttachedLocked(c *Card) bool {
	if c == nil || c.AttachedTo.Kind != TargetCard || c.AttachedTo.ID == uuid.Nil {
		return false
	}
	if findCardOnBattlefield(g, c.AttachedTo.ID) < 0 {
		return false
	}
	return printsReconfigure(c)
}

// AttachedToACreatureForEffect reports whether the named battlefield
// permanent is attached to a creature on the battlefield right now:
// CR 702.151a's "Activate only if this permanent is attached to a
// creature", and The Reality Chip's "as long as ~ is attached to a
// creature". Effective characteristics.
//
// Caller must hold g.mu.
func (g *Game) AttachedToACreatureForEffect(cardID uuid.UUID) bool {
	c := findBattlefieldCard(g, cardID)
	if c == nil || c.AttachedTo.Kind != TargetCard {
		return false
	}
	host := findBattlefieldCard(g, c.AttachedTo.ID)
	return host != nil && host.IsCreature()
}

// reconfiguredEffect is one attached reconfigure Equipment's CR
// 702.151b effect: in layer 4 it is not a creature and has no creature
// types.
type reconfiguredEffect struct {
	target    uuid.UUID
	timestamp int64
}

func (e reconfiguredEffect) Layer() (Layer, SubLayer) { return Layer4Type, 0 }
func (e reconfiguredEffect) Timestamp() int64         { return e.timestamp }
func (e reconfiguredEffect) AppliesTo(target *Card, _ *Game) bool {
	return target != nil && target.InstanceID == e.target
}
func (e reconfiguredEffect) Apply(c *Characteristic, _ *Card, _ *Game) {
	c.loseCreatureType()
}
func (e reconfiguredEffect) RemovesAbilities() bool      { return false }
func (e reconfiguredEffect) ContinuesAfterRemoval() bool { return false }

// loseCreatureType removes the creature card type and every creature
// subtype (CR 205.3d), and with them "is every creature type". Other
// types and subtypes are kept.
func (c *Characteristic) loseCreatureType() {
	types := c.Types[:0:0]
	for _, t := range c.Types {
		if !equalFoldASCII(t, "Creature") {
			types = append(types, t)
		}
	}
	c.Types = types
	subtypes := c.Subtypes[:0:0]
	for _, s := range c.Subtypes {
		if !IsCreatureType(s) {
			subtypes = append(subtypes, s)
		}
	}
	c.Subtypes = subtypes
	c.AllCreatureTypes = false
}

// reconfigureContinuousEffectsLocked is the layer pass's source list for
// CR 702.151b: one effect per attached reconfigure Equipment, ordered
// at its attach. A link with no attach time (a restore point that
// predates it) orders at the permanent's own timestamp.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) reconfigureContinuousEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !g.reconfiguredAttachedLocked(c) {
			continue
		}
		ts := c.AttachedAt
		if ts == 0 {
			ts = c.layerTimestamp()
		}
		out = append(out, reconfiguredEffect{target: c.InstanceID, timestamp: ts})
	}
	return out
}

// UnattachSourceForEffect is the second half of CR 702.151a, "Unattach
// this permanent", as the ability resolves. The source must still be
// the permanent whose ability this is (CR 400.7): one that left, or
// came back as a new object, is not unattached, and that is a quiet
// EventAttachSkipped, as for AttachSourceForEffect. A source that is
// no longer attached to anything is a no-op.
//
// Caller must hold g.mu.
func (g *Game) UnattachSourceForEffect(item *StackItem) error {
	if item == nil {
		return ErrInvalidParam
	}
	if g.AbilitySourceGoneForEffect(item) {
		g.emitAttachSkippedLocked(item.SourceCardID, TargetRef{},
			"the ability's source is no longer the permanent it was activated from")
		return nil
	}
	return g.UnattachForEffect(item.SourceCardID)
}
