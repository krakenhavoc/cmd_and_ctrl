package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elder Mastery — Enchantment — Aura {3}{U}{B}{R} (Conflux, #5587):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has flying.
//	 Whenever enchanted creature deals damage to a player, that player
//	 discards two cards."
//
// Drake Umbra's two statics and a damage trigger on the Aura: any damage
// the enchanted creature deals to any player, combat or not. The damaged
// player chooses the two cards (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f99f44b4-a88f-403b-bbc8-bf1a45557124",
		Name:         "Elder Mastery",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ConB2EnchantedDealtDamageToAPlayer,
				"Elder Mastery — that player discards two cards",
				effects.ConB2ThatPlayerDiscards(2, "Elder Mastery — discard two cards")),
		},
	})
}
