package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammerheim Deadeye — Creature — Giant Warrior, {3}{R}, 3/3:
//
//	"Echo {5}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, destroy target creature with flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6962ccc-9883-48ca-9a5f-aeb68e66fd97",
		Name:         "Hammerheim Deadeye",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Hammerheim Deadeye", "{5}{R}"),
			effects.Targeting(effects.WhenThisEnters("Hammerheim Deadeye — destroy target creature with flying", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
		},
	})
}
