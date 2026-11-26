package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aspect of Gorgon — Enchantment — Aura {2}{B} (Journey into Nyx):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+3 and has deathtouch."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b65276e-7d03-4700-834c-48178fe98353",
		Name:         "Aspect of Gorgon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 3),
			effects.GrantToAttached("deathtouch"),
		},
	})
}
