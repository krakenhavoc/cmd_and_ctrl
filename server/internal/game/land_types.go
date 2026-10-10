package game

import (
	"strings"

	"github.com/google/uuid"
)

// land_types.go is CR 205.3i's list of land types, and the one subtype
// operation that needs it: CR 305.7's "sets a land's subtype to one or
// more of the basic land types" (ADR 0109 §1, #1881).
//
// The engine had no list. landTypeMana knew the five basic land types
// for their mana, AllCreatureTypes knew CR 205.3m's creature types, and
// the other twelve land types appeared only as literals in card files.
// A type SET has to know which subtypes are land types, because it
// replaces those and nothing else (CR 205.1a: "the new subtype(s)
// replaces any existing subtypes from the appropriate set"). So a Dryad
// Arbor under Magus of the Moon is a Mountain Dryad, not a Mountain.

// LandTypes is CR 205.3i's list, in its order: "The land types are
// Cave, Desert, Forest, Gate, Island, Lair, Locus, Mine, Mountain,
// Plains, Planet, Power-Plant, Sphere, Swamp, Tower, Town, and Urza's."
// Read-only.
var LandTypes = []string{
	"Cave", "Desert", "Forest", "Gate", "Island", "Lair", "Locus", "Mine",
	"Mountain", "Plains", "Planet", "Power-Plant", "Sphere", "Swamp",
	"Tower", "Town", "Urza's",
}

// BasicLandTypes is CR 305.6's five, in the order the rule and every
// "basic land type of your choice" prompt print them: Plains, Island,
// Swamp, Mountain, Forest. Read-only.
var BasicLandTypes = []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}

// IsLandType reports whether s is one of CR 205.3i's land types. Case is
// ignored, as it is everywhere the engine compares a subtype, and so is
// the apostrophe's shape: Scryfall prints "Urza's" with a straight one
// and the rules text with a curly one.
func IsLandType(s string) bool {
	return typeIn(s, LandTypes)
}

// IsBasicLandType reports whether s is one of CR 305.6's five.
func IsBasicLandType(s string) bool {
	return typeIn(s, BasicLandTypes)
}

func typeIn(s string, list []string) bool {
	if s == "Urza’s" {
		s = "Urza's"
	}
	for _, t := range list {
		if equalFoldASCII(s, t) {
			return true
		}
	}
	return false
}

// SetLandSubtypes is CR 305.7's subtype half: the object's land types
// are replaced by `types`, and every other subtype it has stays (CR
// 205.1a, "from the appropriate set"). "Is every creature type" stays
// too, because that is a creature-type fact (CR 205.3m) and this
// replaces land types only. A type named twice is written once.
//
// It is not SetSubtypes, which replaces EVERY subtype, and which is
// right for "is an Elk creature". The loss of the land's rules-text
// abilities is the same effect's other half: the caller's static or
// record declares RemovesAbilities in layer 4.
func (c *Characteristic) SetLandSubtypes(types []string) {
	kept := make([]string, 0, len(c.Subtypes)+len(types))
	for _, s := range c.Subtypes {
		if !IsLandType(s) {
			kept = append(kept, s)
		}
	}
	for _, t := range types {
		if !typeListHas(kept, t) {
			kept = append(kept, t)
		}
	}
	c.Subtypes = kept
}

// LoseLandTypes is "loses all land types" (ADR 0109 §2, #1604): every
// subtype that is one of CR 205.3i's land types goes, and every other
// subtype stays (CR 205.1a). SetLandSubtypes with nothing to set. The
// card types, the supertypes and the abilities are untouched; the
// intrinsic mana abilities go with the basic land types they come from
// (CR 305.6). Ultima's record (ModLoseLandTypes), Lithoform Blight's and
// Alpine Moon's static (effects.LosesAllLandTypes).
func (c *Characteristic) LoseLandTypes() {
	c.SetLandSubtypes(nil)
}

// LandTypeEffect is one resolved effect changing a permanent's land
// types, as the table is shown it (ADR 0109 §1 decision 7): "Island
// until end of turn — Tidal Warrior". The view stamps one per live
// record on the permanent. The type line already shows the result;
// this says why and for how long.
type LandTypeEffect struct {
	// Types are the land types the effect gives, in its order.
	Types []string
	// InAddition is "in addition to its other types" (ModAddSubtypes,
	// CR 205.1b): the land keeps what it had.
	InAddition bool
	// LosesAll is "loses all land types" (ModLoseLandTypes, ADR 0109
	// §2): the land has none of its own left. Types is empty.
	LosesAll bool
	// LosesAbilities is the same effect's "and abilities"
	// (ModLoseAllAbilities in the record that removes the land types).
	// A CR 305.7 set takes the rules-text abilities too, but says so by
	// being a set, so this is false for one.
	LosesAbilities bool
	// Gains are the texts of the abilities the same effect gives the
	// land ("{T}: Add {C}."), in the record's order: Ultima's "and has
	// '{T}: Add {C}.'"
	Gains []string
	// Until is the duration in words ("until end of turn"), or "" for
	// an effect with none (CR 611.2a).
	Until string
	// Source is the name of the card whose effect it is.
	Source string
}

// LandTypeEffectsForEffect lists the live resolved effects that set,
// add or remove a land type on the permanent `cardID`, oldest first:
// every setBasicLandTypes record, every addSubtypes record that adds a
// land type (Navigator's Compass, Sealock Monster, The Legend of
// Kyoshi), and every loseLandTypes record (Ultima, with the rest of
// its effect: the abilities lost and the one gained). A static
// ability's type change (Spreading Seas, Blood Moon, Lithoform Blight)
// is not listed: its source is on the battlefield, and the attachment
// or the card says it.
//
// Caller must hold g.mu (read or write).
func (g *Game) LandTypeEffectsForEffect(cardID uuid.UUID) []LandTypeEffect {
	var out []LandTypeEffect
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if e.Scope != ScopeNone {
			continue
		}
		var types, gains []string
		inAddition, losesAll, losesAbilities := false, false, false
		for _, m := range e.Mods {
			switch m.Kind {
			case ModSetBasicLandTypes:
				types = append(types, m.Subtypes...)
			case ModAddSubtypes:
				for _, s := range m.Subtypes {
					if IsLandType(s) {
						types = append(types, s)
						inAddition = true
					}
				}
			case ModLoseLandTypes:
				losesAll = true
			case ModLoseAllAbilities:
				losesAbilities = true
			case ModGrantAbilities:
				for _, k := range m.Grants {
					if text := GrantTextFor(k); text != "" {
						gains = append(gains, text)
					}
				}
			}
		}
		if (len(types) == 0 && !losesAll) || !scopedAffectsLiveObjectLocked(g, *e, cardID) {
			continue
		}
		effect := LandTypeEffect{
			Types:      types,
			InAddition: inAddition,
			Until:      g.durationPhraseLocked(e.Duration, e.SourceName),
			Source:     scopedEffectDisplayName(e),
		}
		if losesAll {
			// The rest of the same effect, so the chip reads as the card
			// prints it: "no land types, no abilities, '{T}: Add {C}.'"
			effect.LosesAll, effect.LosesAbilities, effect.Gains = true, losesAbilities, gains
		}
		out = append(out, effect)
	}
	return out
}

// durationPhraseLocked is a duration in the words a card prints it in,
// for a chip: "until end of turn", "until Bob's next turn", "for as
// long as Gaea's Liege remains on the battlefield". It is "" for a
// duration with no words (CR 611.2a, "if no duration is stated") and
// for one it has no words for.
//
// Caller must hold g.mu (read or write).
func (g *Game) durationPhraseLocked(d Duration, sourceName string) string {
	switch d.Kind {
	case UntilEndOfTurn:
		return "until end of turn"
	case UntilYourNextTurn:
		return "until " + g.playerNameLocked(d.Player) + "'s next turn"
	case UntilEndOfCombat:
		return "until end of combat"
	case ForAsLongAs:
		// ADR 0109 §2 and §3: every condition of the conjunction, joined
		// as the card prints them ("for as long as you control Seasinger
		// and Seasinger remains tapped").
		var parts []string
		for _, c := range d.conditions() {
			p := conditionPhrase(c, d.CounterKind, sourceName)
			if p == "" {
				return ""
			}
			parts = append(parts, p)
		}
		return "for as long as " + strings.Join(parts, " and ")
	}
	return ""
}

// conditionPhrase is one ForAsLongAs condition in words, or "" when it
// has none (a condition about the source with no source name).
func conditionPhrase(c DurationCondition, counterKind, sourceName string) string {
	switch c {
	case WhilePinnedHasCounter:
		return "it has " + counterArticle(counterKind) + " counter on it"
	case WhilePinnedRemainsTapped:
		return "that permanent remains tapped"
	}
	if sourceName == "" {
		return ""
	}
	switch c {
	case WhileSourceOnBattlefield:
		return sourceName + " remains on the battlefield"
	case WhileYouControlSource:
		return "you control " + sourceName
	case WhileSourceRemainsTapped:
		return sourceName + " remains tapped"
	case WhilePinnedPowerAtMostSource:
		return "its power remains at most " + sourceName + "'s"
	case WhileSourceAttachedToPinned:
		return sourceName + " remains attached to it"
	}
	return ""
}

// counterArticle is "a flood", "an awakening": the counter's kind with
// the article a card prints before it.
func counterArticle(kind string) string {
	if kind != "" && strings.ContainsRune("aeiou", rune(kind[0])) {
		return "an " + kind
	}
	return "a " + kind
}
