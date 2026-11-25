package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tooth Collector — Creature — Human Rogue {2}{B}, 3/2 (Shadows over
// Innistrad, #5543):
//
//	"When this creature enters, target creature an opponent controls gets
//	 -1/-1 until end of turn.
//	 Delirium — At the beginning of each opponent's upkeep, if there are
//	 four or more card types among cards in your graveyard, target creature
//	 that player controls gets -1/-1 until end of turn."
//
// The upkeep trigger is an intervening if (CR 603.4) on the Collector's
// controller's own graveyard. Its target is chosen as the trigger goes on
// the stack, from the creatures of the player whose upkeep it is; with none
// the trigger is removed. A target gone by resolution gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7c165a4-53a2-4117-a452-b9a8fe309c09",
		Name:         "Tooth Collector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Tooth Collector — target creature an opponent controls gets -1/-1 until end of turn", effects.S82OldB283MinusOneOnFirstTarget("Tooth Collector")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
			effects.WithTargetsFrom(
				effects.InterveningIf(effects.B24CondDelirium{},
					effects.On(game.EventBeginUpkeep, effects.AllOf(effects.ByAnOpponent, effects.YouHaveDeliriumWhen),
						"Tooth Collector — delirium: target creature that player controls gets -1/-1 until end of turn", effects.S82OldB283MinusOneOnFirstTarget("Tooth Collector"))),
				effects.S82OldB283CreatureOfUpkeepPlayer),
		},
	})
}
