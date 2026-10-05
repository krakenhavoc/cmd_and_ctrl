package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tireless Tracker — Creature — Human Scout {2}{G} (3/2):
//
//	"Landfall — Whenever a land you control enters, investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")
//	 Whenever you sacrifice a Clue, put a +1/+1 counter on this
//	 creature."
//
// The Clue is the shared template, so a nontoken Clue counts for the
// second ability too: EventSacrifice fires before the zone move, so
// the sacrificed Clue is still there to be read (see Lonis). Cracking
// a freshly made Clue therefore pays out the draw and the counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6555cca-e608-4c87-b835-9cd47fd1f543",
		Name:         "Tireless Tracker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Tireless Tracker — investigate (landfall)", Do(CreateToken{Template: ClueToken(), N: 1})),
			On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Tireless Tracker — put a +1/+1 counter on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
