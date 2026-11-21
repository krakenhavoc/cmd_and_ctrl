package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_bns_b29_helpers.go — shared body for the Foundations Jumpstart
// creatures that care about "another" creature of their own type.

// bnsB29AnotherYouControl is "you control another <word>" read for the
// source's controller through the layers (a changeling counts). The word
// is read as a word of the card's text (CR 612.2).
func bnsB29AnotherYouControl(g *game.Game, source *game.Card, word string) bool {
	return MshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype(word))
}

// bnsB29SelfWhileAnother is the AppliesTo of "as long as you control
// another <word>, this creature gets/has ...": the source alone.
func BnsB29SelfWhileAnother(word string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && bnsB29AnotherYouControl(g, source, word)
	}
}
