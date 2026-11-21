package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Twisted Embrace — Enchantment — Aura {2}{B}{B} (Neon Dynasty, #5493):
//
//	"Enchant artifact or creature you control
//	 When this Aura enters, destroy target creature or planeswalker an
//	 opponent controls.
//	 As long as enchanted permanent is a creature, it gets +1/+1."
//
// Two targets, on two moments: the Aura spell's own enchant clause (an
// artifact or creature its caster controls, rerun as CR 704.5m's check) and
// the enters trigger's destroy target, chosen when the trigger is put on the
// stack. The +1/+1 waits for the host to be a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e29efee-1eed-4bef-8574-0f6ff99d3cda",
		Name:         "Twisted Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.B43PumpAttachedWhen(effects.Cmdr2B19AttachedIsCreature, 1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Twisted Embrace — destroy target creature or planeswalker an opponent controls",
					effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target creature or planeswalker an opponent controls", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())),
		},
	})
}
