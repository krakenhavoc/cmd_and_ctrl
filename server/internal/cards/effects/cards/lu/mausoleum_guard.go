package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mausoleum Guard — Creature — Human Scout {3}{W}, 2/2 (Innistrad, #5572):
//
//	"When this creature dies, create two 1/1 white Spirit creature
//	 tokens with flying."
//
// Doomed Traveler's death trigger at two tokens: the 1/1 white flying
// Spirits come from the shared token table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64f3af46-f8ab-4ba4-9dff-7412cefb4eea",
		Name:         "Mausoleum Guard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Mausoleum Guard — create two 1/1 white Spirit creature tokens with flying", effects.Do(effects.CreateToken{
				Template: effects.TokenCard("1/1 white Spirit with flying"),
				N:        2,
			})),
		},
	})
}
