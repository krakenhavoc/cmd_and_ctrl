package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cavalry Drillmaster — Creature — Human Knight {1}{W}, 2/1
// (Core Set 2019, #5527):
//
//	"When this creature enters, target creature gets +2/+0 and gains first strike until end of turn. (It deals combat damage before creatures without first strike.)"
//
// An enters trigger whose body is the shared pumpAndGrantFirstTarget:
// +2/+0 and first strike until end of turn on the still-legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d89bbfe7-e7ef-4f1c-a6b5-6d8ef4079daa",
		Name:         "Cavalry Drillmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Cavalry Drillmaster — target creature gets +2/+0 and gains first strike until end of turn",
				effects.PumpAndGrantFirstTarget(2, 0, "Cavalry Drillmaster — +2/+0 and first strike", "first strike")),
				effects.TargetCreature("target creature")),
		},
	})
}
