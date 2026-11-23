package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wyll, Blade of Frontiers — Legendary Creature — Human Warlock {1}{R}, 1/1 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"If you would roll one or more dice, instead roll that many dice plus one and ignore
//	 the lowest roll.
//	 Whenever you roll one or more dice, put a +1/+1 counter on Wyll.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// The first line is Pixie Guide's replacement on the dice roll's window (CR 614.1a, CR
// 706.6): one more die is rolled and the lowest ignored, so an ignored roll never
// happened and triggers nothing. The trigger is once per roll of one or more dice (the
// roll-batch shape of Farideh, Devil's Chosen), with Wyll getting one +1/+1 counter. Choose
// a Background is a deck-building rule the deck validator applies (deck/partner.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4fb254cf-b60c-4a3e-b753-4213cb572837",
		Name:         "Wyll, Blade of Frontiers",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.RollOneMoreDieAndIgnoreTheLowest("Wyll, Blade of Frontiers")},
		Triggered: []game.TriggeredAbility{
			effects.AfrWheneverYouRollDiceDo("Wyll, Blade of Frontiers — put a +1/+1 counter on Wyll",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
