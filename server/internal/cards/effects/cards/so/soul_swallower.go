package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Swallower — Creature — Wurm {2}{G}{G}, 3/3 (Shadows over Innistrad,
// #5543):
//
//	"Trample
//	 Delirium — At the beginning of your upkeep, if there are four or more
//	 card types among cards in your graveyard, put three +1/+1 counters on
//	 this creature."
//
// The upkeep trigger is an intervening if (CR 603.4) on the controller's
// own graveyard: checked as the upkeep begins and again as the trigger
// resolves. The counters go on this creature as it is now (a Wurm that left
// and came back is a new object, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cc99047-91dc-4ef2-99f3-65932b82ca31",
		Name:            "Soul Swallower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.B24CondDelirium{},
				effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByYou, effects.YouHaveDeliriumWhen),
					"Soul Swallower — delirium: put three +1/+1 counters on this creature", effects.PlusOneCountersOnThis(3))),
		},
	})
}
