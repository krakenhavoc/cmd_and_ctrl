package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Swarm, Being of Bees — Legendary Creature — Insect Villain {2}{B},
// 2/2:
//
//	"Flash
//	 Flying
//	 Mayhem {B} (You may cast this card from your graveyard for {B} if
//	 you discarded it this turn. Timing rules still apply.)"
//
// Flash is the spell's own timing, kept when it is cast from the
// graveyard (CR 702.187b: "timing rules still apply"). Mayhem (CR
// 702.187, #4628). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "10c8cf3b-2530-49e7-b38e-95cdaae8a6e7",
		Name:             "Swarm, Being of Bees",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash", "flying"},
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{B}")},
	})
}
