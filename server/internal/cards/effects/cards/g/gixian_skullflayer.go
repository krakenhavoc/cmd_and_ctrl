package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gixian Skullflayer — Creature — Phyrexian Human Assassin {2}{B}, 2/3
// (The Brothers' War):
//
//	"At the beginning of your upkeep, if there are three or more creature
//	 cards in your graveyard, put a +1/+1 counter on this creature."
//
// One upkeep trigger with an intervening "if" (CR 603.4), asked as the
// upkeep begins and again as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b2a59a3-cf87-48fe-99e2-4c2063119b55",
		Name:         "Gixian Skullflayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondThreeCreatureCardsInYourGraveyard{},
				effects.AtYourUpkeep("Gixian Skullflayer — put a +1/+1 counter on it",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
