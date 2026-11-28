package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampiric Embrace — Enchantment — Aura {2}{B}{B} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying.
//	 Whenever a creature dealt damage by enchanted creature this turn dies, put
//	 a +1/+1 counter on that creature."
//
// "That creature" is the enchanted creature (the dead one cannot take a
// counter), as Scythe of the Wretched's condition reads the damage record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7a50ece-4fe4-4e54-9319-b92a4144445d",
		Name:         "Vampiric Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB605DealtDamageByEquippedCreatureDied,
				"Vampiric Embrace — put a +1/+1 counter on the enchanted creature", effects.S82OldB725PutCounterOnEnchanted),
		},
	})
}
