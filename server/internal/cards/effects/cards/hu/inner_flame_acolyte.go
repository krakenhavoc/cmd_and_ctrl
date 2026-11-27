package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inner-Flame Acolyte — Creature — Elemental Shaman {1}{R}{R}, 2/2 (Lorwyn, S82
// OLD, #5593):
//
//	"When this creature enters, target creature gets +2/+0 and gains haste until
//	 end of turn.
//	 Evoke {R}"
//
// Battleflight Eagle's pump-and-grant trigger and Mulldrifter's evoke: an
// evoked Acolyte's trigger still resolves before the sacrifice (CR 702.74a), and
// the Acolyte itself is a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "699812af-4dba-49dc-987d-67e472a68971",
		Name:             "Inner-Flame Acolyte",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{R}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Inner-Flame Acolyte — target creature gets +2/+0 and gains haste until end of turn",
					effects.TrekB41PumpAndGrantEachTarget(2, 0, "Inner-Flame Acolyte — +2/+0 and haste", "haste")),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "haste"))),
		},
	})
}
