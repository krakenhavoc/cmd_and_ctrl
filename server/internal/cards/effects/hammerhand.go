package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hammerhand — Enchantment — Aura {R}:
//
//	"Enchant creature
//	 When this Aura enters, target creature can't block this turn.
//	 Enchanted creature gets +1/+1 and has haste."
//
// The enters trigger is War Squeak's with any creature as its target
// (the printed clause does not say "an opponent controls").
//
// No simplification.
func init() {
	const label = "Hammerhand — target creature can't block this turn"
	Register(Spec{
		OracleID:     "8efa0026-c3f8-44e0-8f55-729f9c30ac93",
		Name:         "Hammerhand",
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
					TargetCreature("target creature")),
				ForTargets(CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
