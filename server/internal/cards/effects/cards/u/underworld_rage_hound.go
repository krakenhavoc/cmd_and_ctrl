package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Rage-Hound — Creature — Elemental Dog {1}{R}, 3/1:
//
//	"Underworld Rage-Hound attacks each combat if able.
//	 Escape—{3}{R}, Exile three other cards from your graveyard.
//	 Underworld Rage-Hound escapes with a +1/+1 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a1418e8d-dea5-43e1-b041-3f6a88f9a6d7",
		Name:             "Underworld Rage-Hound",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{3}{R}", 3, 1)},
		Static:           []game.StaticAbility{effects.AttacksEachCombat()},
	})
}
