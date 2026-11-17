package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0b2a59a3-cf87-48fe-99e2-4c2063119b55",
		Name:         "Gixian Skullflayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondThreeCreatureCardsInYourGraveyard{},
				AtYourUpkeep("Gixian Skullflayer — put a +1/+1 counter on it",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
