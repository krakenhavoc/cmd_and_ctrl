package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5d13aa6a-5a04-45e0-bb33-daa707ffaf9f",
		Name:         "Gaea's Courser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondThreeCreatureCardsInYourGraveyard{},
				WheneverThisAttacks("Gaea's Courser — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
