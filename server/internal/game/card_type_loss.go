package game

// Losing a card type loses the subtypes that went with it (#2041).
//
// CR 205.1a: "If an object's card type is removed, the subtypes
// correlated with that card type will remain if they are also the
// subtypes of a card type the object currently has; otherwise, they are
// also removed for the entire time the object's card type is removed."
// Weeping Angel's ruling is the same sentence for one card: while it
// isn't a creature it "continues to be an artifact but will stop being
// both an Alien and an Angel until it becomes a creature again."
//
// # One place, every type-changing effect
//
// The layer pass applies the rule, not the effects. Every layer-4
// application (applyOneEffectLocked, and the CR 613.8 dependency
// probe's applyRaw) runs through applyTypeChange, which notes which
// subtype lists the object's card types allowed before the effect and
// which they allow after it, and drops the subtypes of every list the
// effect closed. So RemoveTypesMod("Creature"), a Theros God's
// devotion clause, impending, reconfigure, a catalog closure writing
// `c.Types = []string{"Enchantment"}`: each loses the subtypes its
// removal took, and none of them has to remember to.
//
// # Timestamps (CR 613.7)
//
// The drop happens in the removing effect's own slot, so the order of
// the bucket decides, as it does for everything else in layer 4:
//
//   - "Isn't a creature" sorted after "becomes a creature": the object
//     is not a creature and has no creature types.
//   - "Isn't a creature" sorted before a later "becomes an artifact
//     creature" (March of the Machines, an animation): the object is a
//     creature again, without the creature types the removal took. CR
//     205.1a removes them "for the entire time the object's card type
//     is removed", and the earlier effect is still removing it; CR
//     205.1b's "retains all of its prior … subtypes" retains what the
//     object has at that point in the layer, which is none. An
//     animation that names its own creature types ("a 3/3 Angel
//     artifact creature") adds those.
//   - When the removal ends, the next pass rebuilds the object from its
//     printed values and the subtypes are back.
//
// # Which lists
//
// A subtype is kept while ANY of the object's remaining card types
// allows its list (typeSetCorrelatesWith): creature types need Creature
// or Kindred (Tribal), land types Land, artifact types Artifact,
// enchantment types Enchantment, planeswalker types Planeswalker, spell
// types Instant or Sorcery, battle types Battle. So an artifact creature
// that stops being a creature keeps Equipment, a Dryad Arbor keeps
// Forest, and a Kindred artifact keeps its creature types. A subtype on
// none of CR 205.3's lists the engine carries is never dropped.
//
// "Is every creature type" (changeling, Maskwood Nexus) is a
// creature-type fact (CR 205.3m), so it goes when the creature list
// closes.
//
// # Not here
//
// CR 205.3d, "an object can't GAIN a subtype that doesn't correspond to
// one of that object's types", is the gaining half and is not enforced
// by the pass: a scoped record is one effect per mod, and a record that
// lists its subtype before its card type ("becomes a Bird creature" as
// AddSubtypes then AddTypes) would lose its own subtype mid-record.
// SetTypesMod's registration check refuses the mismatch it can see.

// subtypeSetMask is one bit per subtypeSet.
type subtypeSetMask uint8

func (m subtypeSetMask) has(s subtypeSet) bool { return m&(1<<s) != 0 }

// correlatedSubtypeSets is the subtype lists an object with card types
// `types` may have subtypes from (CR 205.3c, 205.3d).
func correlatedSubtypeSets(types []string) subtypeSetMask {
	var m subtypeSetMask
	for _, t := range types {
		switch {
		case equalFoldASCII(t, "Creature"), equalFoldASCII(t, "Kindred"), equalFoldASCII(t, "Tribal"):
			m |= 1 << subtypeSetCreature
		case equalFoldASCII(t, "Land"):
			m |= 1 << subtypeSetLand
		case equalFoldASCII(t, "Artifact"):
			m |= 1 << subtypeSetArtifact
		case equalFoldASCII(t, "Enchantment"):
			m |= 1 << subtypeSetEnchantment
		case equalFoldASCII(t, "Instant"), equalFoldASCII(t, "Sorcery"):
			m |= 1 << subtypeSetSpell
		case equalFoldASCII(t, "Battle"):
			m |= 1 << subtypeSetBattle
		case equalFoldASCII(t, "Planeswalker"):
			m |= 1 << subtypeSetPlaneswalker
		}
	}
	return m
}

// dropSubtypesOfLostTypes is CR 205.1a on one characteristic after an
// effect changed its card types: every subtype whose list `before`
// allowed and the current card types do not is removed, and so is "is
// every creature type" when the creature list closed. A new slice is
// written, so a cached baseline is never written through.
func (c *Characteristic) dropSubtypesOfLostTypes(before subtypeSetMask) {
	lost := before &^ correlatedSubtypeSets(c.Types)
	if lost == 0 {
		return
	}
	if lost.has(subtypeSetCreature) {
		c.AllCreatureTypes = false
	}
	if len(c.Subtypes) == 0 {
		return
	}
	var kept []string
	for i, s := range c.Subtypes {
		if !lost.has(subtypeSetOf(s)) {
			if kept != nil {
				kept = append(kept, s)
			}
			continue
		}
		if kept == nil {
			kept = make([]string, i, len(c.Subtypes))
			copy(kept, c.Subtypes[:i])
		}
	}
	if kept != nil {
		c.Subtypes = kept
	}
}

// applyTypeChange runs one layer-4 effect on one object with CR 205.1a
// applied to what it removed.
func applyTypeChange(eff ContinuousEffect, c *Characteristic, target *Card, g *Game) {
	before := correlatedSubtypeSets(c.Types)
	eff.Apply(c, target, g)
	c.dropSubtypesOfLostTypes(before)
}
