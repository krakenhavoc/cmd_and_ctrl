package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// See Red — Enchantment — Aura {1}{R} (Rivals of Ixalan, S82 OLD batch 217,
// #5531):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1 and has first strike.
//	 At the beginning of your end step, if you didn't attack with a creature this
//	 turn, sacrifice this Aura."
//
// Curious Obsession's sacrifice clause (the shared constructor): an intervening
// "if" (CR 603.4) checked as your end step begins and again as it resolves,
// with "you" the Aura's controller. The pump and the first strike are the
// Aura's statics.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd33af3e-ab8c-486c-baae-a37cc857f180",
		Name:         "See Red",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 1),
			effects.GrantToAttached("first strike"),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepSacrificeThisIfYouDidntAttack("See Red — sacrifice this Aura (you didn't attack with a creature this turn)"),
		},
	})
}
