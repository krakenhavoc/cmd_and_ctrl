package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Remedy — Enchantment {2}{B} (Magic Origins, S82 older sets):
//
//	"If an opponent would gain life, that player loses that much life
//	 instead."
//
// Plague Drone's Rot Fly on an enchantment: a CR 614 replacement on the life
// an opponent would GAIN (LifeGainBecomes, opponents' scope), so a catalog
// gain, a drain's gain half and lifelink alike become a loss of the same
// amount. The controller's own gains are untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51b114cd-c174-4d74-a854-07c68b41fc9d",
		Name:         "Tainted Remedy",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return -n },
				Scope:  effects.LifeGainByOpponents,
				Label:  "Tainted Remedy: an opponent loses that much life instead",
			}.Build(),
		},
	})
}
