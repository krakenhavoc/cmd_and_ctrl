package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Squeak — Enchantment — Aura {R} (Bloomburrow, #4467):
//
//	"Enchant creature
//	 When this Aura enters, target creature an opponent controls can't
//	 block this turn.
//	 Enchanted creature gets +1/+1 and has haste."
//
// The enchant clause and the two statics are the shared Aura shapes
// (EnchantCreature, PumpAttached, GrantToAttached). The entry trigger is
// Summit Intimidator's "can't block this turn", with the target narrowed
// to a creature an opponent controls and declared as CantBlock.
//
// No simplification.
func init() {
	const label = "War Squeak — target creature an opponent controls can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "54702a84-3125-41a7-a875-fe2ec870836d",
		Name:         "War Squeak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WhenThisEnters(label, effects.TargetCreatureCantBlockThisTurn(label)),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
