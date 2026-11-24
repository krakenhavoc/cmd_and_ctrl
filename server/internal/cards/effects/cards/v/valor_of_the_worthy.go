package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valor of the Worthy — Enchantment — Aura {W} (Kaldheim, #5505):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1.
//	 When enchanted creature leaves the battlefield, create a 1/1 white
//	 Spirit creature token with flying."
//
// Every departure counts (dies, exile, bounce), not only death: the
// leaves-the-battlefield event is harvested while the Aura is still
// attached (see EquippedCreatureDied), so the attachment names the
// creature that left. The token is the Aura's controller's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "84534792-7834-4a3c-9940-c1eccbe6e9cf",
		Name:         "Valor of the Worthy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.EnchantedPermanentLeft,
				"Valor of the Worthy — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 1})),
		},
	})
}
