package game

import (
	"regexp"
	"strings"
)

// split.go — split cards (CR 709), ADR 0103.
//
// A split card has two halves on one card (CR 709.1). ADR 0034 gave
// every multi-face card a face list and made the flat printed fields
// (Name, TypeLine, ManaCost, Colors, …) the materialisation of the
// face that is up. A split card is the one layout for which "the face
// that is up" depends on WHERE the card is:
//
//	stack              the half being cast (CR 709.3b) — SetFace(i),
//	                   unchanged — or both halves for a fused spell
//	                   (CR 702.102b, 709.4d)
//	battlefield        a Room: the halves whose door is unlocked
//	                   (CR 709.5) — see rooms.go
//	every other zone   both halves combined (CR 709.4)
//
// So a split card has two more materialisers beside SetFace, both
// here, and they keep the ADR 0034 bet: every reader of name, cost,
// colour and type is right without learning that split cards exist.
// MoveCard calls them on the way into a zone; the cast path calls
// SetFace or materialiseFused for the stack; the door writers call
// materialiseDoors.
//
// ActiveFace stays 0 for the whole card, so CatalogKey keys a split
// card in hand on its LEFT half's entry — which is what every
// face-walking reader (the view's per-face stamps, the enumerator,
// CardCastableFromAnyFace) already overrides with SetFace on a copy.

// splitNameSeparator joins the two names of a split card, the way
// Scryfall and every decklist print them ("Fire // Ice").
const splitNameSeparator = " // "

// IsSplitCard reports whether c is a two-halved split card the split
// rules apply to. A CR 722.3c prepare copy is excluded: its
// characteristics are the prepare spell's alone (ADR 0090), and its
// layout is `prepare` anyway.
func IsSplitCard(c Card) bool { return isSplitCard(&c) }

// isSplitCard is IsSplitCard without the copy.
func isSplitCard(c *Card) bool {
	return c.Layout == LayoutSplit && len(c.Faces) == 2 && !c.PrepareCopy
}

// HasSharedTypeLine is CR 709.5's test: a split PERMANENT card whose
// halves share one type line — a Room. Structural rather than a check
// for the Room subtype, because the rule is written about the shape.
// All 30 Rooms in the Scryfall dump pass it and no other card does.
//
// False for a face-down object: a face-down Room is a 2/2 with no
// text (CR 708.2a) and has no halves to lock or unlock.
func HasSharedTypeLine(c Card) bool { return hasSharedTypeLine(&c) }

// hasSharedTypeLine is HasSharedTypeLine without the copy, for
// catalogKeyOf.
func hasSharedTypeLine(c *Card) bool {
	if !isSplitCard(c) || c.faceDownPermanent() {
		return false
	}
	if c.Faces[0].TypeLine != c.Faces[1].TypeLine {
		return false
	}
	_, types, _ := ParseTypeLine(c.Faces[0].TypeLine)
	for _, t := range types {
		switch strings.ToLower(t) {
		case "artifact", "creature", "enchantment", "land", "planeswalker", "battle":
			return true
		}
	}
	return false
}

// NamesOf is every name the object has right now (CR 709.4a, 201.4b):
// "an object has the chosen name if one of its names is the chosen
// name". One name for an ordinary card, two for a split card in hand
// or a fused spell, and for a Room on the battlefield the names of its
// unlocked halves — none at all when both are locked (CR 709.5).
//
// Read off the materialised Name, which the split materialisers write
// as the halves joined by " // ", so the answer is right in every zone
// without this function knowing which zone the card is in.
func NamesOf(c Card) []string {
	if c.Name == "" {
		return nil
	}
	if !IsSplitCard(c) {
		return []string{c.Name}
	}
	return strings.Split(c.Name, splitNameSeparator)
}

// HasName reports whether one of the object's names is `name`
// (CR 709.4a).
func HasName(c Card, name string) bool {
	for _, n := range NamesOf(c) {
		if n == name {
			return true
		}
	}
	return false
}

// SettleImported materialises a freshly imported card the way a card
// OUT of play is shown: face 0 for every multi-face card (ADR 0034),
// and for a split card both halves combined (CR 709.4, ADR 0103). The
// deck importer's one call, so an imported split card never shows its
// left half alone in a hand or a library.
func (c *Card) SettleImported() {
	c.SetFace(0)
	c.materialiseSplitWhole()
}

// materialiseSplitWhole writes a split card's COMBINED characteristics
// into the flat printed fields (CR 709.4): both names, the two mana
// costs concatenated (so its colours and mana value are the combined
// ones, CR 709.4b, 202.3d), and every type on either half (CR 709.4c).
// The state of a split card in every zone but the stack and the
// battlefield, and of a fused spell on the stack (CR 709.4d).
//
// A no-op for anything that is not a split card.
func (c *Card) materialiseSplitWhole() {
	if !IsSplitCard(*c) {
		return
	}
	c.materialiseHalves(true, true)
	c.TypeLine = combinedTypeLine(c.Faces[0].TypeLine, c.Faces[1].TypeLine)
}

// materialiseHalves writes the named halves' names, costs and colours
// into the flat fields, left first. Neither half present is CR 709.5's
// fully locked Room: no name, no mana cost and so no colour. The type
// line is the caller's, because what it is depends on why the caller
// is asking.
func (c *Card) materialiseHalves(left, right bool) {
	var names []string
	cost := ""
	var colors []string
	for i, on := range []bool{left, right} {
		if !on {
			continue
		}
		f := c.Faces[i]
		names = append(names, f.Name)
		cost += f.ManaCost
		colors = unionColors(colors, f.Colors)
	}
	c.ActiveFace = 0
	c.Name = strings.Join(names, splitNameSeparator)
	c.ManaCost = cost
	c.Colors = colors
	c.Power = 0
	c.Toughness = 0
	c.VariableToughness = false
	c.StartingLoyalty = 0
	c.StartingDefense = 0
	c.effective = nil
}

// combinedTypeLine is the type line an object with both halves' types
// has (CR 709.4c): the shared line when the halves share one (every
// Room, most split instants), otherwise every supertype, type and
// subtype of either half in printed order — Commit // Memory is an
// "Instant Sorcery" in hand.
func combinedTypeLine(a, b string) string {
	if a == b {
		return a
	}
	sa, ta, ua := ParseTypeLine(a)
	sb, tb, ub := ParseTypeLine(b)
	return composeTypeLine(unionStrings(sa, sb), unionStrings(ta, tb), unionStrings(ua, ub))
}

// unionStrings is a ∪ b, a's order first, without duplicates.
func unionStrings(a, b []string) []string {
	out := append([]string(nil), a...)
	for _, s := range b {
		dup := false
		for _, have := range out {
			if strings.EqualFold(have, s) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, s)
		}
	}
	return out
}

// unionColors is a ∪ b in WUBRG order.
func unionColors(a, b []string) []string {
	have := map[string]bool{}
	for _, x := range a {
		have[strings.ToUpper(x)] = true
	}
	for _, x := range b {
		have[strings.ToUpper(x)] = true
	}
	var out []string
	for _, x := range []string{"W", "U", "B", "R", "G"} {
		if have[x] {
			out = append(out, x)
		}
	}
	return out
}

// --- aftermath (CR 702.127) -------------------------------------------

// A keyword line at the start of the face's text or of one of its
// lines. The importer drops both keywords from Card.Keywords (neither
// is a canonical engine keyword), so the face's own oracle text is the
// one place the card says it — and it is on every imported face.
var (
	aftermathLine = regexp.MustCompile(`(?m)^Aftermath\b`)
	fuseLine      = regexp.MustCompile(`(?m)^Fuse\b`)
)

// AftermathFace is the index of the half that prints aftermath, or -1.
// Always the right half in the dump (27 cards), but read rather than
// assumed.
func AftermathFace(c Card) int {
	if !IsSplitCard(c) {
		return -1
	}
	for i, f := range c.Faces {
		if aftermathLine.MatchString(f.OracleText) {
			return i
		}
	}
	return -1
}

// isAftermathHalf reports whether c, as materialised, is the aftermath
// half of a split card — the half CR 702.127a lets its owner cast from
// a graveyard and from nowhere else.
func isAftermathHalf(c Card) bool {
	if c.Fused {
		return false
	}
	af := AftermathFace(c)
	return af >= 0 && c.ActiveFace == af && c.Name == c.Faces[af].Name
}

// aftermathZoneRule is CR 702.127a's zone half, asked of a card
// materialised as the half being cast:
//
//   - "This half of this split card can't be cast from any zone other
//     than a graveyard" — refused from anywhere else, under any
//     permission (ErrCastZoneNotAllowed).
//   - "You may cast this half of this split card from your graveyard"
//     — opens the graveyard for that half and no other (`opens`).
//
// The other half is untouched: it casts from hand like any card, and
// from a graveyard only under some other permission (a Snapcaster
// Mage flashback grant opens it; aftermath never does).
func aftermathZoneRule(c Card, zone ZoneKind) (opens bool, err error) {
	if !isAftermathHalf(c) {
		return false, nil
	}
	if zone != ZoneGraveyard {
		return false, ErrCastZoneNotAllowed
	}
	return true, nil
}

// FaceCastableFromZone reports whether face `face` of c may be cast out
// of `zone` as far as the card's own split rules go — false only for an
// aftermath half anywhere but a graveyard (CR 702.127a). The view reads
// it to publish no announce surface for a half the zone forbids.
func FaceCastableFromZone(c Card, face int, zone ZoneKind) bool {
	probe := c
	probe.SetFace(face)
	_, err := aftermathZoneRule(probe, zone)
	return err == nil
}

// --- fuse (CR 702.102) ------------------------------------------------

// HasFuse reports whether the split card has fuse (CR 702.102a).
func HasFuse(c Card) bool {
	if !IsSplitCard(c) {
		return false
	}
	for _, f := range c.Faces {
		if fuseLine.MatchString(f.OracleText) {
			return true
		}
	}
	return false
}

// fusedKeySuffix marks the synthetic catalog key of a fused split
// spell. The two halves register under "<oracle>" and "<oracle>#1"
// (ADR 0034 §5); a fused spell is both at once, so its key names the
// oracle ID and says "both" (fusedCatalogDef builds its definition
// from the two). It cannot collide with a face key, whose suffix is a
// number.
const fusedKeySuffix = "#fused"

// FusedCatalogKey is the catalog key of a fused cast of this card.
func FusedCatalogKey(oracleID string) string {
	if oracleID == "" {
		return ""
	}
	return oracleID + fusedKeySuffix
}

// materialiseFused turns a copy of a split card with fuse into a
// fused split spell (CR 702.102a-b): both halves' combined
// characteristics, and the Fused mark that CatalogKey reads.
func (c *Card) materialiseFused() {
	if !IsSplitCard(*c) {
		return
	}
	c.materialiseSplitWhole()
	c.Fused = true
}
