package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Recruiter — 1/1 Creature — Human Advisor for {2}{R}:
//
//	"When this creature enters, search your library for a creature
//	 card with power 2 or less, reveal it, put it into your hand,
//	 then shuffle."
//
// A tutor with a power ceiling, which in this deck means it finds
// Ragavan, Marauding Mako, Siren Stormtamer or the commander. The
// power check reads the printed value: a card in the library has no
// battlefield characteristics for the layer engine to modify, so
// there's nothing else it could mean.
//
// The ETB is a triggered ability and uses the stack (#578). It used to
// run from the direct AsEnters hook, which gave nobody a response
// window; now the trigger waits for every player to pass, like every
// other "When ~ enters".
func init() {
	Register(Spec{
		OracleID:     "4d6a1391-817a-4ddc-840d-886b138eeb3f",
		Name:         "Imperial Recruiter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Imperial Recruiter — search for a creature with power 2 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.Power <= 2
					},
					Dest:    game.ZoneHand,
					Limit:   1,
					Reveal:  true,
					Shuffle: true,
					Reason:  "Imperial Recruiter — a creature with power 2 or less",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
