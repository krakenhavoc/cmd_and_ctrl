package game

import "fmt"

// set_creature_types.go is "<it> becomes a <creature type>" from a
// resolved effect (#2144): Polymorphist's Jest's "becomes a blue Frog",
// Frodo, Sauron's Bane's "it becomes a Halfling Scout". CR 205.1a: "when
// an effect makes an object's subtype a new one, the new subtype
// replaces any existing subtypes appropriate for that card type", so
// the creature's old creature types go and the new ones are its only
// ones.
//
// It is not ModSetTypes. That one is The Irencrag's "becomes a
// legendary Equipment artifact" and also SETS the card types, so an
// artifact creature under it would stop being an artifact. These cards
// name only a creature type: the object keeps its card types, its
// supertypes and its subtypes of other kinds (an Equipment creature
// stays an Equipment), which is how the rulings on Polymorphist's Jest
// and its kind ("it retains any other card types or supertypes") read
// it.

// ModSetCreatureTypes replaces the object's creature types with
// Subtypes (layer 4, CR 205.1a, 613.1d). Reads Subtypes, at least one,
// each a creature type. "Is every creature type" goes too (CR 205.3m):
// changeling's, which the baseline carries before the layer pass, and
// an AllCreatureTypesMod with an earlier timestamp. One with a later
// timestamp applies on top and wins (CR 613.7).
//
// An object that has no card type with creature types when this
// applies (a Vehicle that has stopped being a creature) gets none (CR
// 205.3d); its other subtypes are untouched.
//
// Not a CR 613.1f removal: a card that also says "loses all abilities"
// carries ModLoseAllAbilities in the same record.
const ModSetCreatureTypes ModKind = "setCreatureTypes"

// SetCreatureTypesMod is "becomes a <creature types>" (layer 4,
// CR 205.1a): Polymorphist's Jest's SetCreatureTypesMod("Frog").
func SetCreatureTypesMod(types ...string) Mod {
	return Mod{Kind: ModSetCreatureTypes, Subtypes: copyStrings(types)}
}

// setCreatureTypesModProblem is the registration and restore check for
// ModSetCreatureTypes: at least one type, and every one a creature
// type. A record naming none would strip every creature type and give
// nothing back, which no card prints. "" when sound.
func setCreatureTypesModProblem(m Mod) string {
	if m.Kind != ModSetCreatureTypes {
		return ""
	}
	if len(m.Subtypes) == 0 {
		return "a setCreatureTypes mod names no creature type"
	}
	for _, s := range m.Subtypes {
		if !IsCreatureType(s) {
			return fmt.Sprintf("a setCreatureTypes mod names %q, which is not a creature type", s)
		}
	}
	return ""
}

// SetCreatureTypes is "is a <creature types>" for a static ability that
// replaces the creature types alone (CR 205.1a, Fog on the Barrow-Downs'
// "enchanted creature is a Spirit"): every creature type goes, the other
// subtypes and every card type stay.
func (c *Characteristic) SetCreatureTypes(types []string) { c.setCreatureTypes(types) }

// setCreatureTypes is ModSetCreatureTypes' Apply on one characteristic:
// every creature type goes, "is every creature type" with them, and the
// new ones are added when the object can have creature types.
func (c *Characteristic) setCreatureTypes(types []string) {
	kept := make([]string, 0, len(c.Subtypes)+len(types))
	for _, s := range c.Subtypes {
		if subtypeSetOf(s) != subtypeSetCreature {
			kept = append(kept, s)
		}
	}
	c.AllCreatureTypes = false
	if typeSetCorrelatesWith(subtypeSetCreature, c.Types) {
		for _, s := range types {
			if !typeListHas(kept, s) {
				kept = append(kept, s)
			}
		}
	}
	c.Subtypes = kept
}
