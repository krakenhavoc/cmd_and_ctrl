package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vedalken Heretic — Creature — Vedalken Rogue {G}{U}, 1/1 (Alara Reborn,
// #5586):
//
//	"Whenever this creature deals damage to an opponent, you may draw a
//	 card."
//
// The trigger Snake Umbra grants, printed on the creature: any damage it
// deals to an opponent of its controller, combat or not, and an optional
// draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "001c6369-df13-427d-89df-718d5c09f382",
		Name:         "Vedalken Heretic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OptionalDraw(On(game.EventDealDamage, snakeUmbraDealtDamageToOpponent,
				"Vedalken Heretic — draw a card", Do(DrawCards{N: 1})),
				"Vedalken Heretic — draw a card?", 1),
		},
	})
}
