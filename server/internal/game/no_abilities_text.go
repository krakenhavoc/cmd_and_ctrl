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

// HasNoAbilities reports whether the permanent has no abilities as it is
// now (S82, Jasmine Boreal of the Seven: "creatures you control with no
// abilities"): no keyword in its layered view, no ability another effect
// granted it, and no printed ability that an ability-removing effect has
// not taken away (CR 113.1, CR 613.1f).
//
// A card off the battlefield has only its printed characteristics (CR
// 109.3), so it answers PrintsNoAbilities. On the battlefield a card the
// importer never read (a token, a fixture) is judged by what the engine
// runs for it: no keyword and no catalog ability key.
func (c Card) HasNoAbilities() bool {
	ch := c.effective
	if ch == nil {
		return PrintsNoAbilities(c)
	}
	if len(ch.Abilities) > 0 || len(ch.GrantedAbilities) > 0 {
		return false
	}
	if ch.AbilitiesRemoved || c.FaceDownIsPermanent() {
		return true
	}
	switch c.NoAbilitiesText {
	case PrintedYes:
		return true
	case PrintedNo:
		return false
	}
	return catalogAbilityKeyOf(&c) == ""
}
