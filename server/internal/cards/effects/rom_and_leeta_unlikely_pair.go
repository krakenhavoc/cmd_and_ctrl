package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rom and Leeta, Unlikely Pair — Legendary Creature — Ferengi Bajoran
// {3}{R}, 4/3 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, create two
//	 tapped Treasure tokens."
//
// "Seek out" is an ability word with no rules meaning. The trigger is
// Atlantean Cavalry's: b41DrewYourSecondCardThisTurn reads the turn's
// draw tally after the draw, so it fires on the second card and no
// other, on any player's turn. The Treasures are the printed ones and
// enter tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b98ab107-3908-4425-89d3-634043bb4d4b",
		Name:         "Rom and Leeta, Unlikely Pair",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Rom and Leeta, Unlikely Pair — create two tapped Treasure tokens",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 2})),
		},
	})
}
