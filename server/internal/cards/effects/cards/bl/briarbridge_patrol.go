package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Briarbridge Patrol — Creature — Human Warrior {3}{G}, 3/3 (Shadows over
// Innistrad, #5543):
//
//	"Whenever this creature deals damage to one or more creatures,
//	 investigate.
//	 At the beginning of each end step, if you sacrificed three or more
//	 Clues this turn, you may put a creature card from your hand onto the
//	 battlefield."
//
// The damage trigger is "one or more": damage dealt to several creatures at
// once is one occurrence (CR 603.2c, OncePerBatch). The end-step trigger is
// an intervening if (CR 603.4) on the tally of Clues its controller has
// sacrificed this turn, tokens included, asked as each end step begins and
// again as the ability resolves. The creature is put onto the battlefield,
// not cast, and the "may" is a real decline.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "784f0ca3-efe1-4e22-aedd-57cba7996fa5",
		Name:         "Briarbridge Patrol",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventDealDamage, effects.ThisDealtDamageToACreature,
				"Briarbridge Patrol — investigate", effects.Do(effects.Investigate{N: 1}))),
			effects.InterveningIfYou("you sacrificed three or more Clues this turn", func(g *game.Game, you uuid.UUID) bool {
				return g.SacrificedWithSubtypeThisTurn(you, "Clue") >= 3
			}, effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Briarbridge Patrol — you may put a creature card from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.Creature(),
						Optional: true,
						Label:    "Briarbridge Patrol — you may put a creature card from your hand onto the battlefield",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
