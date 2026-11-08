package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ccb1b532-92b8-4ebf-b3c2-fb5924b116f5",
		Name:         "Slimy Dualleech",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(AtBeginningOfYourCombat(label, pumpAndGrantFirstTarget(1, 0, label, "deathtouch")),
				TargetCreature("target creature you control with power 2 or less", YouControl(), PowerLE(2))),
				ForTargets(BenefitUntilEOT(0, 1, 0, "deathtouch"))),
		},
	})
}
