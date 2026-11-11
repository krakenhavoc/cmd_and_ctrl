package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warbriar Blessing — Enchantment — Aura {1}{G} (Wilds of Eldraine
// Commander, #5368):
//
//	"Enchant creature you control
//	 When this Aura enters, enchanted creature fights up to one target
//	 creature you don't control.
//	 Enchanted creature gets +0/+2."
//
// Pitiless Fists' shape: the enters trigger targets "up to one" creature
// (chosen as the trigger goes on the stack) and the enchanted creature,
// read live as it resolves, fights it. With no target, or if either has
// left, nothing fights (CR 701.14b). The +0/+2 is a layer 7c pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a58954a-1d20-48b8-847c-3615de67e122",
		Name:         "Warbriar Blessing",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(0, 2)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Warbriar Blessing — enchanted creature fights up to one target creature you don't control",
					enchantedCreatureFightsTargetCreature),
				TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1)),
		},
	})
}
