package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloister Gargoyle — Artifact Creature — Gargoyle {2}{W}, 0/4:
//
//	"When this creature enters, venture into the dungeon.
//	 As long as you've completed a dungeon, this creature gets +3/+0
//	 and has flying."
//
// The static reads Player.CompletedDungeons; the layer pass bumps on
// EventDungeonCompleted, so the Gargoyle grows the moment a dungeon is
// completed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1e84b14-c521-4eb7-92ed-bde1131e8662",
		Name:         "Cloister Gargoyle",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WhenThisEntersVenture("Cloister Gargoyle")},
		Static: []game.StaticAbility{
			effects.PumpWhileCompletedADungeon(effects.ThisPermanent, 3, 0),
			effects.KeywordWhileCompletedADungeon(effects.ThisPermanent, "flying"),
		},
	})
}
