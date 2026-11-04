package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mouser Foundry — {1}{R} Artifact (Teenage Mutant Ninja Turtles, #4457):
//
//	"When this artifact enters or leaves the battlefield, create a 1/1
//	 colorless Robot artifact creature token.
//	 {4}{R}, Sacrifice this artifact: It deals 3 damage to target
//	 creature."
//
// The trigger is Subterranean Shambler's enters-or-leaves ability; leaving
// is any exit and looks back in time (CR 603.10a), so sacrificing the
// Foundry to its own ability makes a second Robot. The damage is Goblin
// Cratermaker's sacrificed source dealing damage on its last-known
// information (CR 113.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "404e958b-ead9-4ec7-b786-7b3d65e29967",
		Name:         "Mouser Foundry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrLeaves("Mouser Foundry — create a 1/1 colorless Robot artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Robot artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    Plus(ManaCost("{4}{R}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
