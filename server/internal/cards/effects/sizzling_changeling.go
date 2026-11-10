package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sizzling Changeling — Creature — Shapeshifter {2}{R}, 3/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 When this creature dies, exile the top card of your library. Until the
//	 end of your next turn, you may play that card."
//
// Changeling is the characteristic-defining keyword (CR 702.73a). The dies
// trigger is Kulrath Zealot's impulse exile (ExileTopNUntilYourNextTurn), the
// controller's own library and window as the Changeling last was, so a
// Changeling that dies on an opponent's turn gives you until the end of your
// next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60ad7aaa-8ce6-4449-8e8d-4ccf925010db",
		Name:            "Sizzling Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Sizzling Changeling — exile the top card of your library; you may play it until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopNUntilYourNextTurn(NewContext(g, item), 1)
				}),
		},
	})
}
