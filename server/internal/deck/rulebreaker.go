package deck

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// rulebreaker.go — S82 (#5948): the Mystery Booster Commander "Rulebreaker"
// commanders. Their printed clause,
//
//	"Rulebreaker — A deck with this commander can have Phyrexian cards of
//	 any color identity and any basic land cards."
//	"Rulebreaker — A deck with this commander can have artifact creature and
//	 Equipment cards of any color identity and any basic land cards."
//
// is a deck-construction exception like deckCopyRule's (CR 113.6n): it is
// read off the commander's own oracle text, so no card is named here. It
// lifts the colour-identity check for the cards it describes (a Phyrexian
// card, an artifact creature card, an Equipment card) and for basic lands
// of any type.

var rulebreakerRule = regexp.MustCompile(`(?i)a deck with this commander can have ([^.]*?) cards? of any color identity and any basic land cards`)

// rulebreakerAllowance is what a commander's Rulebreaker clause lets a deck
// hold outside its colour identity.
type rulebreakerAllowance struct {
	basics bool
	// exempt reports whether a card is one of the described cards.
	exempt func(c cards.Card) bool
}

// rulebreakerFor reads a commander's clause, on any face. ok is false when
// it has none.
func rulebreakerFor(c cards.Card) (rulebreakerAllowance, bool) {
	texts := []string{c.OracleText}
	for _, f := range c.CardFaces {
		texts = append(texts, f.OracleText)
	}
	for _, t := range texts {
		m := rulebreakerRule.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		// "Phyrexian" or "artifact creature and Equipment": every card type or
		// subtype word group is one clause joined by "and".
		var groups []string
		for _, part := range strings.Split(strings.ToLower(m[1]), " and ") {
			groups = append(groups, strings.TrimSpace(part))
		}
		return rulebreakerAllowance{basics: true, exempt: func(card cards.Card) bool {
			line := strings.ToLower(card.TypeLine)
			for _, grp := range groups {
				words := strings.Fields(grp)
				all := len(words) > 0
				for _, w := range words {
					if !strings.Contains(line, w) {
						all = false
						break
					}
				}
				if all {
					return true
				}
			}
			return false
		}}, true
	}
	return rulebreakerAllowance{}, false
}

// rulebreakerAllows reports whether any of the deck's commanders lets c
// skip the colour-identity check.
func rulebreakerAllows(commanders []cards.Card, c cards.Card) bool {
	for _, cmd := range commanders {
		a, ok := rulebreakerFor(cmd)
		if !ok {
			continue
		}
		if a.basics && isBasicLand(c) {
			return true
		}
		if a.exempt(c) {
			return true
		}
	}
	return false
}
