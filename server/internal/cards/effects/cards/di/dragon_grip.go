package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dragon Grip — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Enchantment — Aura {2}{R}: Ferocious — If you control a creature with power 4 or
//	greater, you may cast this spell as though it had flash. Enchant creature.
//	Enchanted creature gets +2/+0 and has first strike.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered). The permission is read as the cast begins (CR 601.3).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e83a750-bf86-468e-8342-2c4eb7700dfd",
		Name:         "Dragon Grip",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		FlashIf: effects.FlashIfYou("if you control a creature with power 4 or greater",
			func(g *game.Game, you uuid.UUID) bool { return effects.YouControlPowerFourOrGreater(g, you) }),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantToAttached("first strike"),
		},
	})
}
