package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Courser — Creature — Centaur Soldier {4}{G}, 4/5 (The Brothers' War):
//
//	"Whenever this creature attacks, if there are three or more creature
//	 cards in your graveyard, draw a card."
//
// Angelic Sell-Sword's attack trigger with an intervening "if" (CR 603.4):
// asked as the Courser is declared as an attacker and again as the ability
// resolves, so a graveyard emptied in response draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d13aa6a-5a04-45e0-bb33-daa707ffaf9f",
		Name:         "Gaea's Courser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondThreeCreatureCardsInYourGraveyard{},
				effects.WheneverThisAttacks("Gaea's Courser — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
