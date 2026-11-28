package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bubble Matrix — Artifact {4} (Weatherlight):
//
//	"Prevent all damage that would be dealt to creatures."
//
// Every creature on the battlefield, combat or not, from any source (CR 615.1).
// Players and planeswalkers are not creatures and take damage as usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50261dca-bfeb-40b3-95b6-9057491ef3ad",
		Name:         "Bubble Matrix",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToACreature,
				Label: "Bubble Matrix — prevent all damage that would be dealt to creatures",
			}),
		},
	})
}
