package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War Machine, Legacy of Iron — {2}{R/W} Legendary Artifact Creature —
// Human Hero 1/3 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 At the beginning of combat on your turn, another target creature you
//	 control gets +X/+0 until end of turn, where X is War Machine's power."
//
// Reckless Stormseeker's targeted beginning-of-combat trigger, with Another
// keeping War Machine himself out of the choices. X is read as the trigger
// resolves (CR 608.2h): War Machine's layered power, or the last-known power
// he had if he has left the battlefield in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae6d1aa3-4ceb-4fc1-a3a6-033afd4b3bce",
		Name:            "War Machine, Legacy of Iron",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("War Machine, Legacy of Iron — another target creature you control gets +X/+0 until end of turn, where X is his power",
					warMachineBoost),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}

// warMachineBoost gives the still-legal target +X/+0.
func warMachineBoost(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BoostUntilEOT{
		Target: target,
		Power:  powerNowOrLastKnown(g, item.SourceCardID),
		Label:  "War Machine, Legacy of Iron — +X/+0",
	}.Apply(ctx)
}
