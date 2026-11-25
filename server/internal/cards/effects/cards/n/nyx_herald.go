package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nyx Herald — Enchantment Creature — Centaur Shaman {2}{G}, 2/3 (Theros
// Beyond Death, S82 OLD, #5514):
//
//	"At the beginning of combat on your turn, target enchanted creature or
//	 enchantment creature you control gets +1/+1 and gains trample until end
//	 of turn."
//
// Eidolon of Inspiration's trigger over the target Kestia's attack trigger
// reads: an enchantment creature, or a creature with an Aura attached
// (effects.Enchanted, a Role included, CR 303.4), judged as the target is
// chosen and again as the trigger resolves. The Herald is itself an
// enchantment creature, so it may target itself. With no such creature the
// trigger is not put on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed347609-15f9-4d57-9523-63a33a9dccbc",
		Name:         "Nyx Herald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.AtBeginningOfYourCombat("Nyx Herald — target enchanted creature or enchantment creature you control gets +1/+1 and gains trample until end of turn",
					effects.TrekB41PumpAndGrantEachTarget(1, 1, "Nyx Herald — +1/+1 and trample", "trample")),
				effects.TargetCreature("target enchanted creature or enchantment creature you control",
					effects.YouControl(), effects.Or(effects.Enchanted(), effects.Enchantment()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "trample"))),
		},
	})
}
