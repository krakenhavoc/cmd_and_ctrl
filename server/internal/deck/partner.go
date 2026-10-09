package deck

import (
	"fmt"
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// partner.go — the partner abilities as deck-construction permissions
// (CR 702.124, #2142, #2874, ADR 0144).
//
// CR 702.124a: "Each partner ability allows you to designate two
// legendary cards as your commander rather than one. Each partner
// ability has its own requirements for those two commanders. The
// partner abilities are: partner, partner—[text], partner with [name],
// choose a Background, and Doctor's companion." Every one of them is
// read here, off Scryfall's oracle text rather than off the catalog,
// because deck construction happens before the game and applies to
// every card, whether or not the engine automates anything else about
// it.
//
// One rule judges a pair (commanderPairViolation): two commanders are
// legal when ONE partner ability pairs them, read in the ability's own
// terms (CR 702.124h–m). CR 702.124f says different abilities never
// combine, so a card with partner and a card with partner with [name]
// are not a pair, and CR 702.124g says a card with two of them uses one
// and never makes a third commander. "Partner with [name]"'s second
// ability, the entry search, is a catalog trigger (effects.PartnerWith).
//
// Companion (CR 702.139) is a different mechanic — a card outside the
// deck, not a second commander — and is still refused at import.

// partnerKind is one of CR 702.124's five partner abilities.
type partnerKind int

const (
	// partnerPlain is "Partner" (CR 702.124h).
	partnerPlain partnerKind = iota + 1
	// partnerText is "Partner—[text]" (CR 702.124i): Friends forever,
	// Character select, Father & son, Survivors.
	partnerText
	// partnerWith is "Partner with [name]" (CR 702.124j).
	partnerWith
	// chooseABackground is "Choose a Background" (CR 702.124k).
	chooseABackground
	// doctorsCompanion is "Doctor's companion" (CR 702.124m).
	doctorsCompanion
)

// partnerAbility is one printed partner ability: its kind, and the
// name or text it carries (partner with's name, partner—[text]'s
// text; empty for the others).
type partnerAbility struct {
	kind partnerKind
	arg  string
}

// Keyword lines as Scryfall prints them, each on its own line, with
// reminder text in parentheses on most printings and without on some
// (Frodo, Adventurous Hobbit). The em dash is U+2014.
const (
	partnerWithPrefix    = "Partner with "
	partnerTextPrefix    = "Partner—"
	partnerKeyword       = "Partner"
	chooseABackgroundKW  = "Choose a Background"
	doctorsCompanionKW   = "Doctor's companion"
	doctorsCompanionKWRt = "Doctor’s companion"
)

// partnerAbilities returns every partner ability a card prints, in
// printed order: its own oracle text, then each face's. Nil for a card
// with none.
func partnerAbilities(c cards.Card) []partnerAbility {
	out := partnerAbilitiesIn(c.OracleText)
	for _, face := range c.CardFaces {
		out = append(out, partnerAbilitiesIn(face.OracleText)...)
	}
	return out
}

// partnerAbilitiesIn reads every keyword line of one oracle text. A
// line is its keyword with the reminder text dropped; anything else on
// the line (a sentence that mentions partner) is not the keyword.
func partnerAbilitiesIn(text string) []partnerAbility {
	var out []partnerAbility
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, " ("); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		switch {
		case strings.HasPrefix(line, partnerWithPrefix):
			if name := strings.TrimSpace(strings.TrimPrefix(line, partnerWithPrefix)); name != "" {
				out = append(out, partnerAbility{kind: partnerWith, arg: name})
			}
		case strings.HasPrefix(line, partnerTextPrefix):
			if t := strings.TrimSpace(strings.TrimPrefix(line, partnerTextPrefix)); t != "" {
				out = append(out, partnerAbility{kind: partnerText, arg: t})
			}
		case line == partnerKeyword:
			out = append(out, partnerAbility{kind: partnerPlain})
		case line == chooseABackgroundKW:
			out = append(out, partnerAbility{kind: chooseABackground})
		case line == doctorsCompanionKW || line == doctorsCompanionKWRt:
			out = append(out, partnerAbility{kind: doctorsCompanion})
		}
	}
	return out
}

// partnerWithNames returns the names a card's "Partner with [name]"
// abilities name, in printed order. Nil for a card without the ability.
func partnerWithNames(c cards.Card) []string {
	return partnerArgs(c, partnerWith)
}

// partnerWithNamesIn is partnerWithNames over one oracle text.
func partnerWithNamesIn(text string) []string {
	var out []string
	for _, a := range partnerAbilitiesIn(text) {
		if a.kind == partnerWith {
			out = append(out, a.arg)
		}
	}
	return out
}

// partnerArgs is the arg of every ability of one kind a card prints.
func partnerArgs(c cards.Card, kind partnerKind) []string {
	var out []string
	for _, a := range partnerAbilities(c) {
		if a.kind == kind {
			out = append(out, a.arg)
		}
	}
	return out
}

// hasPartnerAbility reports whether a card prints an ability of `kind`.
func hasPartnerAbility(c cards.Card, kind partnerKind) bool {
	return slices.ContainsFunc(partnerAbilities(c), func(a partnerAbility) bool { return a.kind == kind })
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
		if slices.Contains(cardNames(other), named) {
			return true
		}
	}
	return false
}

// typeLines is the card's type line and each face's.
func typeLines(c cards.Card) []string {
	out := []string{c.TypeLine}
	for _, face := range c.CardFaces {
		out = append(out, face.TypeLine)
	}
	return out
}

// splitTypeLine splits "Legendary Enchantment — Background" into its
// supertypes and card types ("Legendary Enchantment") and its
// subtypes ("Background").
func splitTypeLine(line string) (types, subtypes string) {
	types, subtypes, _ = strings.Cut(line, "—")
	return strings.TrimSpace(types), strings.TrimSpace(subtypes)
}

// isBackgroundCard is CR 702.124k's "legendary Background enchantment
// card": a legendary enchantment with the Background subtype, on the
// card or on one of its faces.
func isBackgroundCard(c cards.Card) bool {
	for _, line := range typeLines(c) {
		types, subtypes := splitTypeLine(line)
		if strings.Contains(types, "Legendary") && strings.Contains(types, "Enchantment") &&
			slices.Contains(strings.Fields(subtypes), "Background") {
			return true
		}
	}
	return false
}

// isLegendaryCreatureCard is "legendary creature card" (CR 702.124m's
// companion half).
func isLegendaryCreatureCard(c cards.Card) bool {
	for _, line := range typeLines(c) {
		types, _ := splitTypeLine(line)
		if strings.Contains(types, "Legendary") && strings.Contains(types, "Creature") {
			return true
		}
	}
	return false
}

// isLoneDoctorCard is CR 702.124m's other card: "a legendary Time Lord
// Doctor creature card that has no other creature types". Time Lord is
// one creature type of two words (CR 205.3m), so the subtypes are
// exactly "Time Lord Doctor".
func isLoneDoctorCard(c cards.Card) bool {
	for _, line := range typeLines(c) {
		types, subtypes := splitTypeLine(line)
		if strings.Contains(types, "Legendary") && strings.Contains(types, "Creature") &&
			strings.Join(strings.Fields(subtypes), " ") == "Time Lord Doctor" {
			return true
		}
	}
	return false
}

// pairedBy reports whether some partner ability lets a and b be the
// deck's two commanders, each in its own terms (CR 702.124h–m). One
// ability does the pairing; abilities never combine (CR 702.124f).
func pairedBy(a, b cards.Card) bool {
	// Partner: both have it.
	if hasPartnerAbility(a, partnerPlain) && hasPartnerAbility(b, partnerPlain) {
		return true
	}
	// Partner—[text]: both have the same one.
	for _, t := range partnerArgs(a, partnerText) {
		if slices.ContainsFunc(partnerArgs(b, partnerText), func(u string) bool { return strings.EqualFold(t, u) }) {
			return true
		}
	}
	// Partner with [name]: each names the other.
	if partnersWith(a, b) && partnersWith(b, a) {
		return true
	}
	// Choose a Background: one of them has it and the other is a
	// legendary Background enchantment card.
	if (hasPartnerAbility(a, chooseABackground) && isBackgroundCard(b)) ||
		(hasPartnerAbility(b, chooseABackground) && isBackgroundCard(a)) {
		return true
	}
	// Doctor's companion: two legendary creature cards, one with the
	// ability and the other a lone Time Lord Doctor.
	companionOf := func(x, doctor cards.Card) bool {
		return hasPartnerAbility(x, doctorsCompanion) && isLegendaryCreatureCard(x) && isLoneDoctorCard(doctor)
	}
	return companionOf(a, b) || companionOf(b, a)
}

// commanderPairViolation judges a deck's two commanders (CR 702.124).
// Nil when a partner ability pairs them. Otherwise one violation that
// says what is wrong:
//
//   - neither card has a partner ability and neither is a Background:
//     CodeTooManyCommanders, as for any two commanders;
//   - otherwise CodeInvalidPartnerPair, naming the card that does not
//     meet the other's requirement.
func commanderPairViolation(a, b cards.Card) *Violation {
	if pairedBy(a, b) {
		return nil
	}
	if len(partnerAbilities(a)) == 0 && len(partnerAbilities(b)) == 0 && !isBackgroundCard(a) && !isBackgroundCard(b) {
		return &Violation{
			Code: CodeTooManyCommanders,
			Message: fmt.Sprintf("deck has 2 commanders, %q and %q; two commanders are allowed only for a partner pair, a commander with Choose a Background and a Background, or a Doctor and their companion",
				a.Name, b.Name),
		}
	}
	offender, msg := pairFault(a, b)
	return &Violation{Code: CodeInvalidPartnerPair, Card: offender, Message: msg}
}

// pairFault names the card at fault in a pair no ability allows, and
// says why, from the point of view of the first card with a
// requirement. Partner with keeps the wording #2142 gave it.
func pairFault(a, b cards.Card) (string, string) {
	prefix := fmt.Sprintf("%q and %q can't be commanders together", a.Name, b.Name)
	// Partner with: name the card whose ability does not name the
	// other one.
	if len(partnerWithNames(a)) > 0 || len(partnerWithNames(b)) > 0 {
		offender, other := a, b
		if partnersWith(a, b) {
			offender, other = b, a
		}
		msg := fmt.Sprintf("%q and %q are not a partner pair: %q has no \"Partner with %s\"",
			a.Name, b.Name, offender.Name, other.Name)
		if named := partnerWithNames(offender); len(named) > 0 {
			msg = fmt.Sprintf("%q and %q are not a partner pair: %q partners with %s, not %q",
				a.Name, b.Name, offender.Name, quoteJoin(named), other.Name)
		}
		return offender.Name, msg
	}
	for _, pair := range [][2]cards.Card{{a, b}, {b, a}} {
		x, y := pair[0], pair[1]
		switch {
		case hasPartnerAbility(x, chooseABackground):
			return y.Name, fmt.Sprintf("%s: %q has Choose a Background, so its second commander has to be a Background, and %q isn't one", prefix, x.Name, y.Name)
		case isBackgroundCard(x):
			return y.Name, fmt.Sprintf("%s: %q is a Background, which can only join a commander with Choose a Background, and %q doesn't have it", prefix, x.Name, y.Name)
		case hasPartnerAbility(x, doctorsCompanion):
			return y.Name, fmt.Sprintf("%s: %q is a Doctor's companion, so its second commander has to be a legendary Time Lord Doctor with no other creature types, and %q isn't one", prefix, x.Name, y.Name)
		case len(partnerArgs(x, partnerText)) > 0:
			return y.Name, fmt.Sprintf("%s: %q has partner—%s, and %q doesn't", prefix, x.Name, partnerArgs(x, partnerText)[0], y.Name)
		case hasPartnerAbility(x, partnerPlain):
			return y.Name, fmt.Sprintf("%s: %q has partner, and %q doesn't", prefix, x.Name, y.Name)
		}
	}
	return b.Name, prefix
}

// loneCommanderViolation judges a deck's only commander against the
// partner rules. A legendary Background enchantment card "can't be
// your commander unless you have also designated a commander with
// 'choose a Background'" (CR 702.124k) — unless it has that ability
// itself (Faceless One). Nil for every other card.
func loneCommanderViolation(c cards.Card) *Violation {
	if !isBackgroundCard(c) || hasPartnerAbility(c, chooseABackground) {
		return nil
	}
	return &Violation{
		Code:    CodeNotLegalCommander,
		Card:    c.Name,
		Message: fmt.Sprintf("%q is a Background, so it can only be a second commander alongside a commander with Choose a Background", c.Name),
	}
}

// backgroundsLast orders a deck's commanders so a Background follows
// the commander that chose it: the command zone, the deck's fallback
// name and the seat's avatar all read the first commander first, and
// that should be the creature (Karlach / Agent of the Iron Throne,
// never the other way round). Stable, so every other pair keeps its
// listed order.
func backgroundsLast(cs []cards.Card) {
	slices.SortStableFunc(cs, func(x, y cards.Card) int {
		return backgroundRank(x) - backgroundRank(y)
	})
}

func backgroundRank(c cards.Card) int {
	if isBackgroundCard(c) && !hasPartnerAbility(c, chooseABackground) {
		return 1
	}
	return 0
}

func quoteJoin(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = fmt.Sprintf("%q", n)
	}
	return strings.Join(q, " and ")
}
