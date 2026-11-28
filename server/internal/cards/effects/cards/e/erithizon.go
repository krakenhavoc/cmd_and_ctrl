package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Erithizon — Creature — Beast {2}{G}{G}, 4/4 (Mercadian Masques, S82 OLD batch
// 698, #5619):
//
//	"Whenever this creature attacks, put a +1/+1 counter on target creature of
//	 defending player's choice."
//
// The Abyss's seam for an attack (OfDefendingPlayersChoice): the Erithizon's
// controller controls the trigger, but the defending player chooses its target,
// a creature of anyone's. The defending player is the player attacked, or the
// controller of the planeswalker or the protector of the battle attacked
// (CR 506.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45645a97-f41d-47fb-b943-293ffc18fc82",
		Name:         "Erithizon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WheneverThisAttacks("Erithizon — put a +1/+1 counter on target creature of defending player's choice", effects.PlusOneCounterOnTheTarget),
			effects.OfDefendingPlayersChoice(effects.TargetCreature("target creature of defending player's choice")),
		)},
	})
}
