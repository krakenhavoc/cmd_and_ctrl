package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaos, the Endless — the back face of Garland, Knight of Cornelia (Final
// Fantasy, #4462): Legendary Creature — Demon, 5/5.
//
//	"Flying
//	 When Chaos dies, put it on the bottom of its owner's library."
//
// A dies trigger whose effect tucks the card from the graveyard to the
// bottom of its owner's library (PutIntoLibrary): the move is the tuck
// route, so a replacement window and a commander's command-zone offer are
// honoured, and a card that left the graveyard in response is a new object
// and is left alone (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        garlandKnightOfCorneliaOracleID + "#1",
		Name:            "Chaos, the Endless",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Chaos, the Endless — put it on the bottom of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					return PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}.Apply(NewContext(g, item))
				}),
		},
	})
}
