package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapwork Cohort — Artifact Creature — Soldier {4}, 3/1:
//
//	"When this creature enters, create a 1/1 colorless Soldier artifact
//	 creature token.
//	 Unearth {2}{W}"
func init() {
	Register(Spec{
		OracleID:     "149694e6-5cf0-4983-9546-a4bd579b2ade",
		Name:         "Scrapwork Cohort",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Scrapwork Cohort — create a 1/1 colorless Soldier artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Soldier artifact"), N: 1})),
		},
		Activated: []ActivatedAbility{Unearth("{2}{W}")},
	})
}
