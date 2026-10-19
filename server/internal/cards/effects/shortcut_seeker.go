package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shortcut Seeker — Creature — Human Rogue {3}{U}, 2/5:
//
//	"Whenever this creature deals combat damage to a player, venture
//	 into the dungeon."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27cfa916-282d-476c-8ee9-ca7f0d84f551",
		Name:         "Shortcut Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(ventureLabel("Shortcut Seeker"), Venture),
		},
	})
}
