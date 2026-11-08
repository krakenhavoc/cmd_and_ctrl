package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Experimental Confectioner — Creature — Human Peasant {2}{B}, 2/3 (Wilds
// of Eldraine, #4471):
//
//	"When this creature enters, create a Food token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	 Whenever you sacrifice a Food, create a 1/1 black Rat creature token
//	 with "This token can't block.""
//
// The entry trigger makes the catalog's Food token (CR 111.10b). The
// sacrifice trigger is Rapacious Guest's (b31YouSacrificedAFood): the
// sacrifice event by the Confectioner's controller, the Food read while
// it is still on the battlefield, so any Food counts, token or not, and
// whatever made you sacrifice it. Its payout is Edgewall Pack's "can't
// block" Rat (CantBlockRatToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87f82d15-5b6b-400f-8a15-d45a2a735f7d",
		Name:         "Experimental Confectioner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Experimental Confectioner — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31YouSacrificedAFood(ev, source, g)
			}, "Experimental Confectioner — create a 1/1 black Rat token that can't block",
				Do(CreateToken{Template: CantBlockRatToken(), N: 1})),
		},
	})
}
