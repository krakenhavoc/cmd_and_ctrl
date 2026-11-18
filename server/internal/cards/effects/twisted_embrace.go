package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2e29efee-1eed-4bef-8574-0f6ff99d3cda",
		Name:         "Twisted Embrace",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{b43PumpAttachedWhen(cmdr2B19AttachedIsCreature, 1, 1)},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Twisted Embrace — destroy target creature or planeswalker an opponent controls",
					destroyFirstLegalCardTarget),
				TargetPermanent("target creature or planeswalker an opponent controls", Or(Creature(), Planeswalker()), OpponentControls())),
		},
	})
}
