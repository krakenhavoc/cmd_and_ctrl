package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slimy Dualleech — Creature — Leech {3}{B}, 2/4 (Murders at Karlov
// Manor, #4469):
//
//	"At the beginning of combat on your turn, target creature you
//	 control with power 2 or less gets +1/+0 and gains deathtouch until
//	 end of turn."
//
// Might of the Ancestors' combat trigger with a narrowed target. The
// power bound is part of the target clause, so a creature whose power
// grew past 2 in response is no longer a legal target and gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Slimy Dualleech — target creature you control with power 2 or less gets +1/+0 and gains deathtouch until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "ccb1b532-92b8-4ebf-b3c2-fb5924b116f5",
		Name:         "Slimy Dualleech",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.AtBeginningOfYourCombat(label, effects.PumpAndGrantFirstTarget(1, 0, label, "deathtouch")),
				effects.TargetCreature("target creature you control with power 2 or less", effects.YouControl(), effects.PowerLE(2))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "deathtouch"))),
		},
	})
}
