package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feast on the Fallen — Enchantment {2}{B} (Jumpstart 2022):
//
//	"At the beginning of each upkeep, if an opponent lost life last turn, put a
//	 +1/+1 counter on target creature you control."
//
// "Last turn" is the previous turn, whoever's it was: the engine stamps the
// players who lost life (damage included) into the next turn's tally as it
// begins (Game.LostLifeLastTurnForEffect). The intervening "if" is checked as
// the upkeep begins and again as the trigger resolves (CR 603.4); the target
// is chosen as it goes on the stack, so with no creature you control there is
// no trigger. A target gone by resolution gets no counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ce9afd9-0c96-42a2-a534-d92ae04f0151",
		Name:         "Feast on the Fallen",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIf(effects.CondOpponentLostLifeLastTurn{},
					effects.On(game.EventBeginUpkeep, effects.OpponentLostLifeLastTurnWhen,
						"Feast on the Fallen — put a +1/+1 counter on target creature you control",
						effects.PlusOneCounterOnTheTarget)),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
