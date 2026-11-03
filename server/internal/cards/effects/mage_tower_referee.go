package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mage Tower Referee — Artifact Creature — Construct {2}, 2/1 (Secrets of
// Strixhaven, #4456):
//
//	"Whenever you cast a multicolored spell, put a +1/+1 counter on this
//	 creature."
//
// A cast trigger on your own spells whose colours, read as the spell sits
// on the stack, number two or more (CR 105.2b; b25CastIsMulticolored, the
// same read as Zenith Chronicler's). A hybrid spell is multicolored
// whatever it was paid with. The counter goes on only while the Referee is
// still the permanent that triggered (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f27988b-0863-4cc1-8eb9-83bd0cd8561c",
		Name:         "Mage Tower Referee",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && b25CastIsMulticolored(ev, g)
			}, "Mage Tower Referee — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
