package game

import (
	"regexp"
	"strconv"
)

// prototype.go — CR 718, prototype cards, and CR 702.160, prototype
// (#5656).
//
//	702.160a Prototype is a static ability that appears on prototype
//	cards that have a secondary set of power, toughness, and mana cost
//	characteristics. A player who casts a spell with prototype can
//	choose to cast that card "prototyped." If they do, the alternative
//	set of its power, toughness, and mana cost characteristics are used.
//	718.3b Both a prototyped spell and the permanent it becomes have
//	only its alternative set of power, toughness, and mana cost
//	characteristics. If that mana cost includes one or more colored
//	mana symbols, the spell and the permanent it becomes are also that
//	color or colors.
//	718.4. In every zone except the stack or the battlefield, and while
//	on the stack or the battlefield when not cast as a prototyped spell,
//	a prototype card has only its normal characteristics.
//
// # A prototype card is a card with two faces
//
// Prototype is not an alternative cost: the prototyped spell's mana
// cost IS the inset cost, so its mana value, its colour and every cost
// that reads "its mana cost" follow, and another alternative cost or a
// cast permission may still apply to it. What the rules describe is a
// second set of printed characteristics chosen as the card is cast,
// kept by the spell and the permanent it becomes, and gone in every
// other zone. That is ADR 0034's face model, so a prototype card is
// one: Faces[0] is the card as printed and Faces[PrototypeFace] the
// same card with the inset's mana cost, power and toughness and the
// colours of that cost. Nothing else differs between the two (CR
// 718.5), and both read ONE catalog entry: the key of a prototyped
// object is the bare oracle ID (underlyingTextKeyOf).
//
// Everything else is the face machinery, unchanged:
//
//   - the choice is the cast's `face` (CastableFaces offers both), made
//     before every announce gate, so the cost, a permission's mana-value
//     filter and the cast restrictions judge the prototyped spell (CR
//     718.3a);
//   - the permanent keeps the face (faceOnResolve);
//   - MoveCard turns the card back to its normal face in every zone but
//     the stack and the battlefield (CR 718.4);
//   - the faces and which one is up are copiable values (CR 718.2a,
//     718.3c, 718.3d), which CopiableValuesOf already carries.
//
// Scryfall's `prototype` layout is a single-faced record whose inset is
// only in the oracle text, so the deck importer builds the faces with
// PrototypeFaces.

// LayoutPrototype is Scryfall's layout for a prototype card (CR 718).
const LayoutPrototype = "prototype"

// PrototypeFace is the face of a prototype card that carries the
// inset's characteristics: the one a prototyped spell and the permanent
// it becomes have up.
const PrototypeFace = 1

// IsPrototypeCard reports whether c is a prototype card with both sets
// of characteristics imported.
func IsPrototypeCard(c Card) bool {
	return c.Layout == LayoutPrototype && len(c.Faces) == 2
}

// Prototyped reports whether this object is a prototyped spell or a
// permanent that was one (CR 718.3b). A copy of either is one too (CR
// 718.3c, 718.3d): it copied the faces and the one that is up.
func (c Card) Prototyped() bool {
	return c.ActiveFace == PrototypeFace && IsPrototypeCard(c)
}

// prototypeLine is the inset as the oracle text prints it, on a line of
// its own: "Prototype {2}{R} — 3/2".
var prototypeLine = regexp.MustCompile(`(?m)^Prototype ((?:\{[^{}\s]+\})+) — (\d+)/(\d+)`)

// ParsePrototype reads the inset's mana cost, power and toughness out
// of a prototype card's oracle text. ok is false for a text with no
// such line.
func ParsePrototype(oracleText string) (cost string, power, toughness int, ok bool) {
	m := prototypeLine.FindStringSubmatch(oracleText)
	if m == nil {
		return "", 0, 0, false
	}
	power, _ = strconv.Atoi(m[2])
	toughness, _ = strconv.Atoi(m[3])
	return m[1], power, toughness, true
}

// PrototypeFaces is the face list of a prototype card: the card as
// printed, then the card with the inset's mana cost, power and
// toughness and the colours of that cost (CR 718.3b, 105.2). `normal`
// is the printed card, oracle text included, since that is where the
// inset is read from. nil when the text has no inset, which leaves the
// card single-faced.
func PrototypeFaces(normal Face) []Face {
	cost, power, toughness, ok := ParsePrototype(normal.OracleText)
	if !ok {
		return nil
	}
	alt := normal
	alt.ManaCost = cost
	alt.Colors = printedColorsFromCost(cost)
	alt.Power = power
	alt.Toughness = toughness
	alt.VariableToughness = false
	return []Face{normal, alt}
}
