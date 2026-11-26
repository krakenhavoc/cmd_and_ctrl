package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skyblinder Staff — Artifact — Equipment {1} (Gatecrash, #5565):
//
//	"Equipped creature gets +1/+0 and can't be blocked by creatures with
//	 flying.
//	 Equip {3}"
//
// The pump is a layer 7c static on the host (PumpAttached). The block rule
// belongs to the Staff and is re-read on every check (Prowler's Helm's
// shape, CantBeBlockedBy on OnAttached): a flying blocker, read through the
// layered view as blockers are declared (CR 509.1b), is refused, and moving
// the Staff moves the rule with it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3fea3d1a-067b-4eae-8433-f7e1e7b724a3",
		Name:         "Skyblinder Staff",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnAttached(), effects.HasKeyword("flying"), "creatures with flying"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
