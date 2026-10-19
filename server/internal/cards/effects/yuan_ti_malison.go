package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yuan-Ti Malison — Creature — Snake Rogue {1}{U}, 2/1:
//
//	"This creature can't be blocked as long as it's attacking alone.
//	 Whenever this creature deals combat damage to a player, venture
//	 into the dungeon."
//
// The evasion is Gutter Skulker's block rule; the venture is
// game.VentureForEffect (#4100).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0e4dda1-2ffb-4353-a292-dcbc966967b4",
		Name:         "Yuan-Ti Malison",
		Completeness: CompletenessFull,
		BlockRules:   []game.BlockRule{CantBeBlockedWhile(OnSelf(), AttackingAlone())},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(ventureLabel("Yuan-Ti Malison"), Venture),
		},
	})
}
