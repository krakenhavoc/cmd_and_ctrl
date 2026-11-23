package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daybreak Combatants — Creature — Human Warrior {2}{R}, 2/2 (Innistrad:
// Crimson Vow):
//
//	"Haste
//	 When this creature enters, target creature gets +2/+0 until end of
//	 turn."
//
// The target may be the Combatants itself or any other creature; it is
// chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbb3f959-ded7-4570-bacf-0e1394dbef80",
		Name:            "Daybreak Combatants",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Daybreak Combatants — target creature gets +2/+0",
				effects.BoostTargetUntilEOT("Daybreak Combatants — +2/+0", 2, 0)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0))),
		},
	})
}
