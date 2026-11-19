package deck

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rulebreaker.go — "Rulebreaker — A deck with this commander can have …"
// (Mystery Booster Commander, S82 GAP batch 3, #5948): a commander's own
// deck-construction exceptions to the colour identity rule (CR 903.4),
// read off its oracle text on any face, like deckCopyLimit, so no card is
// named in the validator.
//
// Three clauses are read, each by its printed words:
//
//	"any land cards"                    Grizzlegom, Hurloon Hero: every land
//	                                    card, of any colour identity.
//	"any basic land cards"              Maular, the Next Evolution: basic lands
//	                                    of any colour identity.
//	"creature cards with mana value 7   Maular: creature cards of mana value 7
//	 or greater of any color identity"  or greater, of any colour identity.
//
// A sentence the reader does not recognise grants nothing: the strict
// direction.

// rulebreakerRule picks out the sentence a deck exception is printed in.
var rulebreakerRule = regexp.MustCompile(`(?i)a deck with this commander can have ([^.]*)\.`)

// rulebreakerMinMV is the mana value Maular's creature clause starts at.
const rulebreakerMinMV = 7

// rulebreaks is what a deck's commanders let it hold past its identity.
type rulebreaks struct {
	anyLand, basicLands, bigCreatures bool
}

// commanderRulebreaks reads the exceptions every commander's text grants.
func commanderRulebreaks(commanders []cards.Card) rulebreaks {
	var r rulebreaks
	for _, c := range commanders {
		texts := []string{c.OracleText}
		for _, f := range c.CardFaces {
			texts = append(texts, f.OracleText)
		}
		for _, t := range texts {
			for _, m := range rulebreakerRule.FindAllStringSubmatch(t, -1) {
				clause := strings.ToLower(m[1])
				if strings.Contains(clause, "any land cards") {
					r.anyLand = true
				}
				if strings.Contains(clause, "any basic land cards") {
					r.basicLands = true
				}
				if strings.Contains(clause, "creature cards with mana value 7 or greater of any color identity") {
					r.bigCreatures = true
				}
			}
		}
	}
	return r
}

// exempts reports whether the card may ignore the commander's colour
// identity.
func (r rulebreaks) exempts(c cards.Card) bool {
	switch {
	case r.anyLand && strings.Contains(c.TypeLine, "Land"):
		return true
	case r.basicLands && isBasicLand(c):
		return true
	case r.bigCreatures && strings.Contains(c.TypeLine, "Creature"):
		total := 0
		for _, cost := range companionManaCosts(c) {
			parsed, err := game.ParseCost(cost)
			if err != nil {
				return false
			}
			total += parsed.ManaValue()
		}
		return total >= rulebreakerMinMV
	}
	return false
}
