package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summit Intimidator — Creature — Yeti {3}{R}, 4/3 (Tarkir: Dragonstorm,
// #4463):
//
//	"Reach
//	 When this creature enters, target creature can't block this turn."
//
// Zog, Triceraton Castaway's entry trigger: any creature may be the
// target, the Intimidator included, and the restriction lasts until the
// end of the turn.
//
// No simplification.
func init() {
	const label = "Summit Intimidator — target creature can't block this turn"
	Register(Spec{
		OracleID:        "b772d35a-d365-47ce-b1dd-a2ba574c74ea",
		Name:            "Summit Intimidator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
					TargetCreature("target creature")),
				ForTargets(CantBlockUntilEOT(0, 0, 0))),
		},
	})
}
