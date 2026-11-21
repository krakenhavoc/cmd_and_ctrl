package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raging Goblinoids — Creature — Goblin Berserker Villain {4}{R}, 5/4
// (Marvel's Spider-Man, #4460):
//
//	"Haste
//	 Mayhem {2}{R} (You may cast this card from your graveyard for {2}{R}
//	 if you discarded it this turn. Timing rules still apply.)"
//
// Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "c807b53f-2917-41a0-812c-57240b623365",
		Name:             "Raging Goblinoids",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{2}{R}")},
	})
}
