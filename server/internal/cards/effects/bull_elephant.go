package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bull Elephant — Creature — Elephant {3}{G}, 4/4 (Multiverse Gift Box,
// #5632):
//
//	"When this creature enters, sacrifice it unless you return two
//	 Forests you control to their owner's hand."
//
// returnLandsOrSacrifice (return_lands_or_sacrifice.go): exactly two
// Forests or none, and the Elephant stays only if both reached a hand.
// "Forests" is the land type, not the name, so a Stomping Ground counts,
// tapped or untapped, and the word is read through the card's text
// (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ed04ece-64d2-41a4-bf36-b993cbbf2aac",
		Name:         "Bull Elephant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			bullElephantReturn.trigger("Bull Elephant — sacrifice it unless you return two Forests you control to their owner's hand"),
		},
	})
}

var bullElephantReturn = newReturnLandsOrSacrifice(
	"Bull Elephant — return two Forests you control to their owner's hand (or sacrifice Bull Elephant)", 2,
	"choose-cards/bull-elephant-return-two-forests", "mgb/bull-elephant-after-return",
	func(source *game.Card, land game.Card) bool {
		forest := "Forest"
		if source != nil {
			forest = source.TextSubtype("Forest")
		}
		return land.HasSubtype(forest)
	})
