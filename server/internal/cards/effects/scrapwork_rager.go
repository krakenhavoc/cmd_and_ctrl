package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapwork Rager — Artifact Creature — Phyrexian Horror {4}, 2/2:
//
//	"When this creature enters, you draw a card and you lose 1 life.
//	 Unearth {3}{B}"
func init() {
	Register(Spec{
		OracleID:     "552867ec-5741-4e34-813f-4aaf185661f7",
		Name:         "Scrapwork Rager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Scrapwork Rager — you draw a card and you lose 1 life", b36DrawAndLoseOne),
		},
		Activated: []ActivatedAbility{Unearth("{3}{B}")},
	})
}
