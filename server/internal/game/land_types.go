package game

import "github.com/google/uuid"

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
	// Until is the duration in words ("until end of turn"), or "" for
	// an effect with none (CR 611.2a).
	Until string
	// Source is the name of the card whose effect it is.
	Source string
}

// LandTypeEffectsForEffect lists the live resolved effects that set or
// add a land type on the permanent `cardID`, oldest first: every
// setBasicLandTypes record, and every addSubtypes record that adds a
// land type (Navigator's Compass, Sealock Monster, The Legend of
// Kyoshi). A static ability's type change (Spreading Seas, Blood Moon)
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
		var types []string
		inAddition := false
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
			}
		}
		if len(types) == 0 || !scopedAffectsLiveObjectLocked(g, *e, cardID) {
			continue
		}
		out = append(out, LandTypeEffect{
			Types:      types,
			InAddition: inAddition,
			Until:      g.durationPhraseLocked(e.Duration, e.SourceName),
			Source:     scopedEffectDisplayName(e),
		})
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
	case ForAsLongAs:
		if sourceName == "" {
			return ""
		}
		switch d.Condition {
		case WhileSourceOnBattlefield:
			return "for as long as " + sourceName + " remains on the battlefield"
		case WhileYouControlSource:
			return "for as long as you control " + sourceName
		case WhileSourceRemainsTapped:
			return "for as long as " + sourceName + " remains tapped"
		}
	}
	return ""
}
