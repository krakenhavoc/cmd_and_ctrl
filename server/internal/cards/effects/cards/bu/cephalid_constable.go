package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cephalid Constable — Creature — Octopus Wizard {1}{U}{U}, 1/1 (Judgment,
// #5611):
//
//	"Whenever this creature deals combat damage to a player, return up to that
//	 many target permanents that player controls to their owner's hand."
//
// The target clause is built as the trigger is put on the stack from the damage
// event: the damaged player's permanents, up to as many as the damage dealt
// (CR 603.3d). Each still-legal target is returned to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5236bf7-c84f-4c64-bf6b-3a89c05e161a",
		Name:         "Cephalid Constable",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WithTargetsFrom(
				effects.WheneverThisDealsCombatDamageToAPlayer("Cephalid Constable — return up to that many target permanents that player controls to their owner's hand", effects.BounceEachLegalTarget),
				effects.OldB623TargetUpToThatManyPermanentsThatPlayerControls),
		},
	})
}
