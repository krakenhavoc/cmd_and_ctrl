package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gibbering Fiend — Creature — Devil {1}{R}, 2/1 (Shadows over Innistrad,
// #5543):
//
//	"When this creature enters, it deals 1 damage to each opponent.
//	 Delirium — At the beginning of each opponent's upkeep, if there are
//	 four or more card types among cards in your graveyard, this creature
//	 deals 1 damage to that player."
//
// The upkeep trigger is an intervening if (CR 603.4) on the Fiend's
// controller's own graveyard, checked as each opponent's upkeep begins and
// again as the trigger resolves; "that player" is the upkeep's player,
// read off the event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2aa03852-6c38-4e91-afed-0c4cd5964598",
		Name:         "Gibbering Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gibbering Fiend — deals 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})),
			effects.InterveningIf(effects.B24CondDelirium{},
				effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, effects.YouHaveDeliriumWhen),
					"Gibbering Fiend — delirium: deals 1 damage to that player", effects.S82OldB282DamageUpkeepPlayer(1))),
		},
	})
}
