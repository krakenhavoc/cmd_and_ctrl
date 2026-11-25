package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mogis's Favor — Enchantment — Aura {B}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/-1.
//	 Escape—{2}{B}, Exile two other cards from your graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:         "312dbc3a-3ffb-4ed2-a278-8950d8ce8c86",
		Name:             "Mogis's Favor",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Aura: game.AuraBenefit},
		Targets:          effects.EnchantCreature(),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{2}{B}", 2)},
		Static:           []game.StaticAbility{effects.PumpAttached(2, -1)},
	})
}
