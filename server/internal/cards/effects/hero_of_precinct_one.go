package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hero of Precinct One — Creature — Human Warrior {1}{W}, 2/2
// (Ravnica Allegiance reprint):
//
//	"Whenever you cast a multicolored spell, create a 1/1 white Human
//	 creature token."
//
// A cast trigger on your own spells whose colours, read as the spell sits
// on the stack, number two or more (CR 105.2b; b25CastIsMulticolored, Mage
// Tower Referee's read). The token is the printed 1/1 white Human.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a66802a-76f9-4a31-a776-24110dcfca64",
		Name:         "Hero of Precinct One",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && b25CastIsMulticolored(ev, g)
			}, "Hero of Precinct One — create a 1/1 white Human creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
		},
	})
}
