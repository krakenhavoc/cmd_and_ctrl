package deck

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// partner.go — "Partner with [name]" as a deck-construction permission
// (#2142, CR 702.124j).
//
// CR 702.124j makes "partner with [name]" two abilities. The first is
// the one this file reads: "You may designate two legendary cards as
// your commander rather than one if each has a 'partner with [name]'
// ability with the other's name." The second, the entry search, is a
// catalog trigger (effects.PartnerWith).
//
// It is read off Scryfall's oracle text, not off the catalog, because
// deck construction happens before the game and applies to every card,
// whether or not the engine automates anything else about it.
//
// The other partner abilities (plain partner, "partner—[text]" such as
// Friends forever, choose a Background, Doctor's companion) are still
// not modelled: Resolve refuses plain partner and partner—[text]
// commanders outright, and a Background or Doctor pairing is two
// commanders that are not a partner-with pair, so Validate refuses it.
// CR 702.124f says the abilities never combine, so a pair is legal
// here only when BOTH cards print partner with and each names the
// other.

// partnerWithPrefix opens the keyword line. Scryfall prints it on its
// own line, with reminder text in parentheses on most printings and
// without on some (Frodo, Adventurous Hobbit).
const partnerWithPrefix = "Partner with "

// partnerWithNames returns the names a card's "Partner with [name]"
// abilities name, in printed order: its own oracle text, then each
// face's. Nil for a card without the ability.
func partnerWithNames(c cards.Card) []string {
	names := partnerWithNamesIn(c.OracleText)
	for _, face := range c.CardFaces {
		names = append(names, partnerWithNamesIn(face.OracleText)...)
	}
	return names
}

// partnerWithNamesIn reads every keyword line of one oracle text. The
// name runs from the keyword to the reminder text's parenthesis, or to
// the end of the line.
func partnerWithNamesIn(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, partnerWithPrefix) {
			continue
		}
		name := strings.TrimPrefix(line, partnerWithPrefix)
		if i := strings.Index(name, " ("); i >= 0 {
			name = name[:i]
		}
		if name = strings.TrimSpace(name); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// withoutPartnerWithLines drops the "Partner with [name]" lines from an
// oracle text, so the unsupported-mechanic sniff does not read the
// word "Partner" in them as plain partner.
func withoutPartnerWithLines(text string) string {
	lines := strings.Split(text, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), partnerWithPrefix) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// cardNames is every name a card answers to for "the other's name":
// its full name and, for a card with faces, each face's.
func cardNames(c cards.Card) []string {
	names := []string{c.Name}
	for _, face := range c.CardFaces {
		if face.Name != "" && face.Name != c.Name {
			names = append(names, face.Name)
		}
	}
	return names
}

// partnersWith reports whether one of c's "Partner with" abilities
// names `other`.
func partnersWith(c cards.Card, other cards.Card) bool {
	for _, named := range partnerWithNames(c) {
		for _, n := range cardNames(other) {
			if named == n {
				return true
			}
		}
	}
	return false
}

// commanderPairViolation judges a deck's two commanders (CR 702.124a,
// 702.124j). Nil when each has a "partner with" ability naming the
// other. Otherwise one violation that says what is wrong:
//
//   - neither prints partner with: CodeTooManyCommanders, as before;
//   - only one names the other, or they name other cards:
//     CodeInvalidPartnerPair, naming the card at fault.
func commanderPairViolation(a, b cards.Card) *Violation {
	aNames, bNames := partnerWithNames(a), partnerWithNames(b)
	if len(aNames) == 0 && len(bNames) == 0 {
		return &Violation{
			Code: CodeTooManyCommanders,
			Message: fmt.Sprintf("deck has 2 commanders, %q and %q; two commanders are allowed only for a \"Partner with\" pair",
				a.Name, b.Name),
		}
	}
	aOK, bOK := partnersWith(a, b), partnersWith(b, a)
	if aOK && bOK {
		return nil
	}
	// Name the card whose ability does not name the other one.
	offender, other := a, b
	if aOK {
		offender, other = b, a
	}
	msg := fmt.Sprintf("%q and %q are not a partner pair: %q has no \"Partner with %s\"",
		a.Name, b.Name, offender.Name, other.Name)
	if named := partnerWithNames(offender); len(named) > 0 {
		msg = fmt.Sprintf("%q and %q are not a partner pair: %q partners with %s, not %q",
			a.Name, b.Name, offender.Name, quoteJoin(named), other.Name)
	}
	return &Violation{Code: CodeInvalidPartnerPair, Card: offender.Name, Message: msg}
}

func quoteJoin(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = fmt.Sprintf("%q", n)
	}
	return strings.Join(q, " and ")
}
