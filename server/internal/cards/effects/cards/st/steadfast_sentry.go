package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steadfast Sentry — {2}{W} Creature — Human Soldier 3/2 (Foundations
// Jumpstart):
//
//	"Vigilance
//	 When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0f676ab-f859-4d4f-abe0-d033ffd3c396",
		Name:            "Steadfast Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Steadfast Sentry — put a +1/+1 counter on target creature you control",
				effects.RfCreatureFCountersOnTarget(1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
