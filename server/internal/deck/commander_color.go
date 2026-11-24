package deck

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// commander_color.go — CR 903.4b for the deck checker (#5948):
//
//	903.4b If a commander has a static ability that causes a player to
//	choose its color before the game begins, that choice applies during
//	deck construction and throughout the game, even as the commander
//	changes zones. That choice affects the commander's color identity.
//
// The Prismatic Piper, Faceless One and Clara Oswald print it. A
// decklist has nowhere to write the choice, so the deck shows it by the
// cards it holds: the colours its cards need beyond the other
// commanders' identity are the chosen colour, as long as they are no
// more colours than the deck has such commanders. The engine reads the
// same deck the same way as the game begins (game/commander_color.go),
// and asks only when the deck shows no choice.

// choosesColorBeforeGame reports whether a commander prints CR 903.4b's
// ability. Read off the printed text, as the partner abilities are
// (partner.go).
func choosesColorBeforeGame(c cards.Card) bool {
	texts := []string{c.OracleText}
	for _, f := range c.CardFaces {
		texts = append(texts, f.OracleText)
	}
	for _, t := range texts {
		if strings.Contains(strings.ToLower(t), "choose a color before the game begins") {
			return true
		}
	}
	return false
}

// commanderAllowedIdentity is the colour identity the deck's commanders
// allow (CR 903.4, the union of a pair's, CR 702.124c), with each CR
// 903.4b commander's chosen colour read off the deck: the colours the
// mainboard and the companion need outside the other commanders'
// identity, when there are no more of them than such commanders. A deck
// that needs more is judged without them, and the colour-identity check
// names its off-colour cards as it always did.
func commanderAllowedIdentity(list *List) map[string]struct{} {
	allowed := map[string]struct{}{}
	choosers := 0
	for _, cmd := range list.Commanders {
		if choosesColorBeforeGame(cmd) {
			choosers++
		}
		for sym := range identitySet(cmd.ColorIdentity) {
			allowed[sym] = struct{}{}
		}
	}
	if choosers == 0 {
		return allowed
	}
	need := map[string]struct{}{}
	deckCards := append([]cards.Card(nil), list.Mainboard...)
	if list.Companion != nil {
		deckCards = append(deckCards, *list.Companion)
	}
	for _, c := range deckCards {
		for _, sym := range offIdentity(c, allowed) {
			need[sym] = struct{}{}
		}
	}
	if len(need) > choosers {
		return allowed
	}
	for sym := range need {
		allowed[sym] = struct{}{}
	}
	return allowed
}
