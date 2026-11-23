package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapwork Cohort — Artifact Creature — Soldier {4}, 3/1:
//
//	"When this creature enters, create a 1/1 colorless Soldier artifact
//	 creature token.
//	 Unearth {2}{W}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "149694e6-5cf0-4983-9546-a4bd579b2ade",
		Name:         "Scrapwork Cohort",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scrapwork Cohort — create a 1/1 colorless Soldier artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Soldier artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{2}{W}")},
	})
}
