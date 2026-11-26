package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cast into Darkness — Enchantment — Aura {1}{B} (Journey into Nyx):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-0 and can't block."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4160e621-3f13-486b-af17-728ce85f7ea5",
		Name:         "Cast into Darkness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-2, 0),
			effects.RestrictAttached(game.CantBlock),
		},
	})
}
