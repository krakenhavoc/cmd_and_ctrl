package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapwork Rager — Artifact Creature — Phyrexian Horror {4}, 2/2:
//
//	"When this creature enters, you draw a card and you lose 1 life.
//	 Unearth {3}{B}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "552867ec-5741-4e34-813f-4aaf185661f7",
		Name:         "Scrapwork Rager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Scrapwork Rager — you draw a card and you lose 1 life", effects.B36DrawAndLoseOne),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{3}{B}")},
	})
}
