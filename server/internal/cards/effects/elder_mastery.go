package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f99f44b4-a88f-403b-bbc8-bf1a45557124",
		Name:         "Elder Mastery",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, conB2EnchantedDealtDamageToAPlayer,
				"Elder Mastery — that player discards two cards",
				conB2ThatPlayerDiscards(2, "Elder Mastery — discard two cards")),
		},
	})
}
