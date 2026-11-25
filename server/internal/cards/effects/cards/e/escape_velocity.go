package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escape Velocity — Enchantment — Aura {R}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has haste.
//	 Escape—{1}{R}, Exile two other cards from your graveyard."
//
// An escaped Aura is cast from the graveyard and so targets as it is cast
// (CR 303.4f); it stays on the battlefield until it leaves.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2db4c197-982b-4c23-88e7-6ddf17378bb4",
		Name:             "Escape Velocity",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Aura: game.AuraBenefit},
		Targets:          effects.EnchantCreature(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{1}{R}", 2)},
		Static:           []game.StaticAbility{effects.PumpAttached(1, 0), effects.GrantToAttached("haste")},
	})
}
