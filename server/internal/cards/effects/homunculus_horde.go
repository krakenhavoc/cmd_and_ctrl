package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Homunculus Horde — Creature — Homunculus {3}{U}, 2/2 (Foundations,
// #4465):
//
//	"Whenever you draw your second card each turn, create a token that's
//	 a copy of this creature."
//
// Atlantean Cavalry's second-draw trigger (b41DrewYourSecondCardThisTurn,
// once a turn) with Scute Swarm's CreateTokenCopy of the source: its
// copiable values (CR 707.2), or its last-known ones if it has left the
// battlefield (CR 608.2h). Each copy has the trigger too, so every Horde
// you control makes a copy on the second draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3c69bf4-880f-4198-a67c-87d23d9c3d41",
		Name:         "Homunculus Horde",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Homunculus Horde — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
