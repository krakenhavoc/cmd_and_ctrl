package game

import "fmt"

// become_named.go is "<it> becomes a <type line> named <name>" from a
// resolved effect (#2562, ADR 0093 amendment 2026-10-10): The Irencrag's
// "you may have The Irencrag become a legendary Equipment artifact named
// Everflame, Heroes' Legacy". Two ScopedEffect mods, each in the layer
// the rules put it in, both in one record so they share one timestamp
// and are one effect (CR 613.7):
//
//   - ModSetName, layer 3. CR 612.8: an effect that sets an object's
//     name is a text-changing effect, and the object "loses any names it
//     had and has only the specified name". The legend rule reads the
//     effective name (legendRuleChoicesLocked), and so does every
//     "named" check and the wire, so the new name is the one they see.
//   - ModSetTypes, layer 4. "Becomes a [supertype] [subtype] [card
//     type]" with no "in addition to its other types": CR 205.1a's set,
//     not CR 205.1b's add.
//
// Neither is copiable (CR 707.2): a copy reads PrintedValues, so a
// Clone of Everflame is a Clone of The Irencrag.

// ModSetName is "named <Text>" (layer 3, CR 612.8). Reads Text, which
// must not be empty.
const ModSetName ModKind = "setName"

// ModSetTypes is "becomes a <Supertypes> <Subtypes> <Types>" (layer 4)
// with CR 205.1a's meaning:
//
//   - The card types become Types. An instant or a sorcery keeps that
//     type (CR 205.1a's one exception).
//   - Each subtype in Subtypes replaces the object's subtypes of the
//     same set (artifact types, creature types, …); subtypes of the other
//     sets stay, unless their card type is gone, in which case they go
//     with it (CR 205.1a, "the subtypes correlated with that card type
//     … are also removed").
//   - Supertypes are GAINED: CR 205.4b, "when an object gains or loses a
//     supertype, it retains any other supertypes it had".
//
// Reads Types (at least one), Subtypes (each one for one of Types, CR
// 205.3d) and Supertypes. Not a CR 613.1f removal: what the object loses
// with its old types is the same effect's ModLoseAllAbilities, in layer
// 6, when the card says so.
const ModSetTypes ModKind = "setTypes"

// SetNameMod is "named <name>" (layer 3, CR 612.8).
func SetNameMod(name string) Mod { return Mod{Kind: ModSetName, Text: name} }

// SetTypesMod is "becomes a <supertypes> <subtypes> <types>" (layer 4,
// CR 205.1a): The Irencrag's "a legendary Equipment artifact" is
// SetTypesMod([]string{"Artifact"}, []string{"Equipment"}, "Legendary").
func SetTypesMod(types, subtypes []string, supertypes ...string) Mod {
	return Mod{Kind: ModSetTypes, Types: copyStrings(types), Subtypes: copyStrings(subtypes), Supertypes: copyStrings(supertypes)}
}

// knownSupertypes is CR 205.4a's list.
var knownSupertypes = []string{"Basic", "Legendary", "Ongoing", "Snow", "World"}

// becomeNamedModProblem is the registration and restore check for the
// two kinds, and for the one field only ModSetTypes reads. "" when the
// mod is sound.
func becomeNamedModProblem(m Mod) string {
	if m.Kind != ModSetTypes && len(m.Supertypes) > 0 {
		return fmt.Sprintf("mod %q names supertypes, which only a setTypes mod reads", m.Kind)
	}
	switch m.Kind {
	case ModSetName:
		if m.Text == "" {
			return "a setName mod names no name"
		}
	case ModSetTypes:
		if len(m.Types) == 0 {
			return "a setTypes mod names no card type"
		}
		for _, s := range m.Subtypes {
			if !typeSetCorrelatesWith(subtypeSetOf(s), m.Types) {
				return fmt.Sprintf("a setTypes mod names the subtype %q, which is not one of its card types' (CR 205.3d)", s)
			}
		}
		for _, s := range m.Supertypes {
			if !typeIn(s, knownSupertypes) {
				return fmt.Sprintf("a setTypes mod names %q, which is not a supertype", s)
			}
		}
	}
	return ""
}

// setTypes is ModSetTypes' Apply on one characteristic. Every slice it
// writes is a new one, so a cached baseline is never written through.
func (c *Characteristic) setTypes(types, subtypes, supertypes []string) {
	next := make([]string, 0, len(types)+1)
	for _, t := range c.Types {
		if equalFoldASCII(t, "Instant") || equalFoldASCII(t, "Sorcery") {
			next = append(next, t)
		}
	}
	for _, t := range types {
		if !typeListHas(next, t) {
			next = append(next, t)
		}
	}
	c.Types = next

	replaced := map[subtypeSet]bool{}
	for _, s := range subtypes {
		replaced[subtypeSetOf(s)] = true
	}
	kept := make([]string, 0, len(c.Subtypes)+len(subtypes))
	for _, s := range c.Subtypes {
		set := subtypeSetOf(s)
		if replaced[set] || !typeSetCorrelatesWith(set, c.Types) {
			continue
		}
		kept = append(kept, s)
	}
	for _, s := range subtypes {
		if !typeListHas(kept, s) {
			kept = append(kept, s)
		}
	}
	c.Subtypes = kept
	// "Is every creature type" is a creature-type fact (CR 205.3m): a
	// new creature type replaces it, and so does losing every type that
	// has creature types.
	if replaced[subtypeSetCreature] || !typeSetCorrelatesWith(subtypeSetCreature, c.Types) {
		c.AllCreatureTypes = false
	}

	if len(supertypes) > 0 {
		super := append([]string(nil), c.Supertypes...)
		for _, s := range supertypes {
			if !typeListHas(super, s) {
				super = append(super, s)
			}
		}
		c.Supertypes = super
	}
}

// subtypeSet is which of CR 205.3's subtype lists a subtype is on.
type subtypeSet uint8

const (
	// subtypeSetOther is a subtype on none of the lists below: in
	// practice a planeswalker type, whose list the engine does not
	// carry. It is never dropped with a card type, which is the
	// conservative reading for a list this file cannot check.
	subtypeSetOther subtypeSet = iota
	subtypeSetCreature
	subtypeSetLand
	subtypeSetArtifact
	subtypeSetEnchantment
	subtypeSetSpell
	subtypeSetBattle
)

// artifactTypes is CR 205.3g's list.
var artifactTypes = []string{
	"Attraction", "Blood", "Bobblehead", "Book", "Clue", "Contraption", "Equipment", "Food",
	"Fortification", "Gold", "Heartwood", "Incubator", "Infinity", "Junk", "Lander", "Map",
	"Mutagen", "Powerstone", "Spacecraft", "Stone", "Treasure", "Vehicle", "Vibranium",
}

// enchantmentTypes is CR 205.3h's list.
var enchantmentTypes = []string{
	"Aura", "Background", "Cartouche", "Case", "Class", "Curse", "Plan", "Role", "Room", "Rune",
	"Saga", "Shard", "Shrine",
}

// spellTypes is CR 205.3k's list.
var spellTypes = []string{"Adventure", "Arcane", "Lesson", "Omen", "Trap"}

func subtypeSetOf(s string) subtypeSet {
	switch {
	case IsCreatureType(s):
		return subtypeSetCreature
	case IsLandType(s):
		return subtypeSetLand
	case typeIn(s, artifactTypes):
		return subtypeSetArtifact
	case typeIn(s, enchantmentTypes):
		return subtypeSetEnchantment
	case typeIn(s, spellTypes):
		return subtypeSetSpell
	case equalFoldASCII(s, "Siege"):
		return subtypeSetBattle
	}
	return subtypeSetOther
}

// typeSetCorrelatesWith reports whether an object with card types
// `types` can have a subtype of `set` (CR 205.3c, 205.3d). Kindred
// shares creature types (CR 205.3m), and "Tribal" is its old name (CR
// 308.3).
func typeSetCorrelatesWith(set subtypeSet, types []string) bool {
	has := func(names ...string) bool {
		for _, n := range names {
			if typeListHas(types, n) {
				return true
			}
		}
		return false
	}
	switch set {
	case subtypeSetCreature:
		return has("Creature", "Kindred", "Tribal")
	case subtypeSetLand:
		return has("Land")
	case subtypeSetArtifact:
		return has("Artifact")
	case subtypeSetEnchantment:
		return has("Enchantment")
	case subtypeSetSpell:
		return has("Instant", "Sorcery")
	case subtypeSetBattle:
		return has("Battle")
	}
	return true
}
