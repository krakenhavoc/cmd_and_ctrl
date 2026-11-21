package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f0e4dda1-2ffb-4353-a292-dcbc966967b4",
		Name:         "Yuan-Ti Malison",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), effects.AttackingAlone())},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(effects.VentureLabel("Yuan-Ti Malison"), effects.Venture),
		},
	})
}
