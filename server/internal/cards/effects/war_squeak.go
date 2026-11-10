package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "54702a84-3125-41a7-a875-fe2ec870836d",
		Name:         "War Squeak",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
					TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
