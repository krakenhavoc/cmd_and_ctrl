package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inspiring Veteran — Creature — Human Knight {R}{W}, 2/2 (Throne of Eldraine, #5516):
//
//	"Other Knights you control get +1/+1."
//
// Marshal of Zhalfir's lord clause (TribalAnthem).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa1a63dd-acb1-465f-8970-667b8d7c57c9",
		Name:         "Inspiring Veteran",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
