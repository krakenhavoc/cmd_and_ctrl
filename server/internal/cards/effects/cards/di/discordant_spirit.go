package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discordant Spirit — Creature — Spirit {2}{B}{R}, 2/2 (Mirage, S82 OLD b805):
//
//	"At the beginning of each end step, if it's an opponent's turn, put a +1/+1
//	 counter on this creature for each 1 damage dealt to you this turn.
//	 At the beginning of your end step, remove all +1/+1 counters from this
//	 creature."
//
// The first line is an intervening "if" (CR 603.4) that the end step is not
// its controller's turn, checked as the step begins and as the trigger
// resolves. It counts the damage actually dealt to the controller this turn,
// combat or not, after prevention (PlayerTurnTally.DamageTaken). The second
// line is Force Bubble's remove-all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce785306-eaf5-42b3-9840-884cd690ae0a",
		Name:         "Discordant Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("it's an opponent's turn", func(g *game.Game, you uuid.UUID) bool {
				return !effects.IsActivePlayer(g, you)
			}, effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Discordant Spirit — put a +1/+1 counter on it for each 1 damage dealt to you this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := g.DamageTakenThisTurn(item.Controller)
					if n <= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(effects.NewContext(g, item))
				})),
			effects.AtYourEndStep("Discordant Spirit — remove all +1/+1 counters from it",
				effects.RemoveAllCountersFromThisEffect(game.CounterPlusOne)),
		},
	})
}
