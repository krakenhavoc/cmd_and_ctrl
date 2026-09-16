package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elanor Gardner — Legendary Creature — Halfling Scout {3}{G}, 2/4
// (EDHREC rank 3643):
//
//	"When Elanor enters, create a Food token.
//	 At the beginning of your end step, if you sacrificed a Food this
//	 turn, you may search your library for a basic land card, put
//	 that card onto the battlefield tapped, then shuffle."
//
// The Food-into-lands Halfling. The entry Food is the real token,
// with its own gain-3 ability. The end-step trigger carries an
// intervening-if — a Food sacrificed by the controller this turn,
// read off the event log back to the turn's upkeep — checked when
// the end step begins and again as the trigger resolves (CR 603.4);
// the "you may" is the ordinary trigger prompt, and the search prompt
// is where the controller picks the basic, which enters tapped
// through the search's own flag. Any Food counts: the token Elanor
// made, a Gingerbrute, a Food eaten to a Gilded Goose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad4c39d6-a8b3-4dd9-816d-09761fc9651d",
		Name:         "Elanor Gardner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elanor Gardner — create a Food", b34CreateTokens(FoodToken, 1)),
			Optional(On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b34YouSacrificedAFoodThisTurn(g, source.Controller)
			}, "Elanor Gardner — search for a basic land, put it onto the battlefield tapped", b34SearchBasicTappedIfSacrificedAFood), "Elanor Gardner — you sacrificed a Food this turn. Search your library for a basic land card?"),
		},
	})
}
