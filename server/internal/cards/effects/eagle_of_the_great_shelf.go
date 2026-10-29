package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eagle of the Great Shelf — Creature — Bird Soldier {4}{W}, 2/5 (The Hobbit,
// HOB):
//
//	"Flying
//	 Whenever this creature attacks, it gets +1/+1 until end of turn for
//	 each other creature you control."
//
// The pump reads the board once, as the trigger resolves, and lands as a
// one-time until-end-of-turn boost (CR 611.2c, Goblin Rabblemaster's shape):
// creatures that arrive or die afterwards do not change it. "Other" leaves
// the Eagle out of the count, and a creature is counted by its effective
// type, so a land animated into a creature counts. Flying rides
// PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb4b4b63-682b-4439-9757-84cca7a27db2",
		Name:            "Eagle of the Great Shelf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Eagle of the Great Shelf — +1/+1 until end of turn for each other creature you control",
				eagleOfTheGreatShelfPump),
		},
	})
}

// eagleOfTheGreatShelfPump counts the other creatures the controller has
// right now and pumps the Eagle that much.
func eagleOfTheGreatShelfPump(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != item.SourceCardID {
			n++
		}
	}
	return BoostUntilEOT{
		Target:    item.SourceCardID,
		Power:     n,
		Toughness: n,
		Label:     "Eagle of the Great Shelf — +1/+1 for each other creature you control",
	}.Apply(NewContext(g, item))
}
