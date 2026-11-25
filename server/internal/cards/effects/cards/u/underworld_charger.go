package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Charger — Creature — Nightmare Horse {2}{B}, 3/3:
//
//	"Underworld Charger can't block.
//	 Escape—{4}{B}, Exile three other cards from your graveyard.
//	 Underworld Charger escapes with two +1/+1 counters on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "39694a3f-536e-41e0-adf6-d49de721aabd",
		Name:             "Underworld Charger",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{4}{B}", 3, 2)},
		Static:           []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
