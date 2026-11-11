package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalisse, Reverent Medium — Legendary Creature — Human Cleric {3}{W}{B},
// 3/4:
//
//	"At the beginning of each end step, create X 1/1 white Spirit creature
//	 tokens with flying, where X is the number of tokens you created this
//	 turn."
//
// "Each end step" is any player's. X is the controller's per-turn token
// tally (game.PlayerTurnTally.TokensCreated), read as the trigger resolves,
// so tokens made by an effect between the trigger and its resolution count
// and Spirits Thalisse makes are not counted until the next turn's tally.
// A turn in which no token was created makes none.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77cf4fb1-593f-4240-96d9-6ae9084baeac",
		Name:         "Thalisse, Reverent Medium",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15EndStepBegan(ev)
			}, "Thalisse, Reverent Medium — create X 1/1 white Spirit creature tokens with flying, where X is the number of tokens you created this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.TurnTallyFor(item.Controller).TokensCreated
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: n}.Apply(NewContext(g, item))
				}),
		},
	})
}
