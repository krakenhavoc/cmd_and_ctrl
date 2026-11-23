package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Syndicate Liquidators — {4}{G} Creature — Orion Pirate 4/4 (Star Trek,
// #4973):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, put two +1/+1
//	 counters on target creature you control."
//
// Landfall is an ability word (CR 207.2c). The target is chosen as the
// trigger goes on the stack and re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2f32de5c-4d85-42bc-b1bf-f762a46e5efc",
		Name:            "Syndicate Liquidators",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Landfall("Syndicate Liquidators — put two +1/+1 counters on target creature you control (landfall)",
				effects.TwoPlusOneCountersOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
