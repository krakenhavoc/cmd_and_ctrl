package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Topplegeist — Creature — Spirit {W}, 1/1 (Shadows over Innistrad,
// #5543):
//
//	"Flying
//	 When this creature enters, tap target creature an opponent controls.
//	 Delirium — At the beginning of each opponent's upkeep, if there are
//	 four or more card types among cards in your graveyard, tap target
//	 creature that player controls."
//
// The upkeep trigger is an intervening if (CR 603.4) on the Spirit's
// controller's own graveyard. Its target is chosen as the trigger goes on
// the stack, from the creatures of the player whose upkeep it is; with none
// the trigger is removed. A target gone by resolution is not tapped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a71f404-e48d-4c86-882e-4e21c8210dea",
		Name:            "Topplegeist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Topplegeist — tap target creature an opponent controls", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.WithTargetsFrom(
				effects.InterveningIf(effects.B24CondDelirium{},
					effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, effects.YouHaveDeliriumWhen),
						"Topplegeist — delirium: tap target creature that player controls", effects.TapTheFirstLegalTarget)),
				effects.S82OldB282CreatureOfUpkeepPlayer),
		},
	})
}
