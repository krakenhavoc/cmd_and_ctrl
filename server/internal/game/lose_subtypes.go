package game

// lose_subtypes.go is "<it> isn't a <creature type>" from a resolved
// effect (#4072): Werewolf Pack Leader's "{3}{G}: Until end of turn, this
// creature has base power and toughness 5/3, gains trample, and isn't a
// Human."

// ModLoseSubtypes removes the named subtypes from the object (layer 4,
// CR 613.1d). Reads Subtypes, at least one; the match ignores case, like
// every subtype test. Card types, supertypes and every other subtype are
// untouched, and an object without the subtype is unchanged.
//
// The kind is an on-disk identity: it is written into snapshots by name
// and is never renamed (additive within schema v9; an older binary
// refuses it as an unknown kind).
//
// One limit: "is every creature type" (changeling, CR 702.73) is a flag,
// not a list, so a changeling that "isn't a Human" is still every
// creature type. No card in the catalog does that today.
const ModLoseSubtypes ModKind = "loseSubtypes"

// LoseSubtypesMod is "isn't a <subtype>" (layer 4, CR 613.1d):
// LoseSubtypesMod("Human"). A card file passes the word through the
// creating object's text, LoseSubtypesMod(ctx.TextSubtype("Human")).
func LoseSubtypesMod(subtypes ...string) Mod {
	return Mod{Kind: ModLoseSubtypes, Subtypes: copyStrings(subtypes)}
}

// loseSubtypesModProblem is the registration and restore check for
// ModLoseSubtypes: at least one subtype, none empty. "" when sound.
func loseSubtypesModProblem(m Mod) string {
	if m.Kind != ModLoseSubtypes {
		return ""
	}
	if len(m.Subtypes) == 0 {
		return "a loseSubtypes mod names no subtype"
	}
	for _, s := range m.Subtypes {
		if s == "" {
			return "a loseSubtypes mod names an empty subtype"
		}
	}
	return ""
}

// loseSubtypes is ModLoseSubtypes' Apply on one characteristic.
func (c *Characteristic) loseSubtypes(lost []string) {
	kept := make([]string, 0, len(c.Subtypes))
	for _, s := range c.Subtypes {
		if !typeListHas(lost, s) {
			kept = append(kept, s)
		}
	}
	c.Subtypes = kept
}
