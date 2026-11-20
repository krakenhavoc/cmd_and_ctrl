package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aspect of Lamprey — Enchantment — Aura {3}{B} (Theros Beyond Death, #5514):
//
//	"Enchant creature you control
//	 When this Aura enters, target opponent discards two cards.
//	 Enchanted creature has lifelink."
//
// Enchant creature you control and the lifelink grant (Felidar Umbra's),
// with HYDRA Infiltration's enters discard: the target opponent picks
// the two cards (CR 701.9a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d3c25bb-5175-4fb1-992f-62a44722f535",
		Name:         "Aspect of Lamprey",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("lifelink")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Aspect of Lamprey — target opponent discards two cards", targetPlayerDiscards(2)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
