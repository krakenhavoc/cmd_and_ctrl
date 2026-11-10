package game

import "strings"

// no_abilities_text.go — "a creature card with no abilities" (Fang-Druid
// Summoner, Rise from the Wreck; #4464).
//
// A card in a library or a graveyard has only its printed characteristics
// (CR 109.3), and a card has no abilities when its rules text prints none:
// a vanilla creature. game.Card carries no oracle text, so the deck importer
// reads the text once and stamps Card.NoAbilitiesText, the way it stamps
// ManaAbilityText; SetFace re-derives it from the face that is up.

// NoAbilitiesInText reports whether every one of the texts (the card's and
// each face's) is empty. A card with any rules text, even a lone keyword,
// has an ability. A card with no texts at all reads as having none to
// speak of, which is the importer's own empty record.
func NoAbilitiesInText(texts ...string) bool {
	for _, t := range texts {
		if strings.TrimSpace(t) != "" {
			return false
		}
	}
	return true
}

// PrintsNoAbilities reports whether the card's printed text, as the face
// that is up prints it, is empty. A card the importer never read
// (PrintedUnknown: a token, a fixture, a restore point older than the
// stamp) is not known to be without abilities, so it is not.
func PrintsNoAbilities(c Card) bool {
	return c.NoAbilitiesText == PrintedYes && !c.FaceDownIsPermanent()
}
