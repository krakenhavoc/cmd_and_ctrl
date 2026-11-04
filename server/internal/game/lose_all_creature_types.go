package game

// lose_all_creature_types.go is "<it> loses all creature types" from a
// resolved effect (#4691): Nameless Inversion's "Target creature gets
// +3/-3 and loses all creature types until end of turn."

// ModLoseAllCreatureTypes removes every creature type from the object
// (layer 4, CR 613.1d) and adds none. It reads no field. "Is every
// creature type" goes too (CR 205.3m): changeling's (CR 702.73a), which
// the baseline carries before the layer pass, and an AllCreatureTypesMod
// with an earlier timestamp. One with a later timestamp applies on top
// and wins (CR 613.7).
//
// Removing a subtype doesn't affect the card types (CR 205.1a): the
// object keeps its card types, its supertypes and its subtypes of other
// kinds (an Equipment creature stays an Equipment). It removes no
// ability, so a changeling keeps the keyword while it lasts and is no
// creature type.
//
// It is not ModSetCreatureTypes, which must name the types it gives and
// refuses a record naming none, and not ModLoseSubtypes, which removes
// only the subtypes it names and leaves the every-creature-type flag.
//
// The kind is an on-disk identity: it is written into snapshots by name
// and is never renamed (additive within the schema version; an older
// binary refuses it as an unknown kind).
const ModLoseAllCreatureTypes ModKind = "loseAllCreatureTypes"

// LoseAllCreatureTypesMod is "loses all creature types" (layer 4,
// CR 613.1d).
func LoseAllCreatureTypesMod() Mod { return Mod{Kind: ModLoseAllCreatureTypes} }

// loseAllCreatureTypes is ModLoseAllCreatureTypes' Apply on one
// characteristic: setCreatureTypes with nothing to add.
func (c *Characteristic) loseAllCreatureTypes() { c.setCreatureTypes(nil) }
