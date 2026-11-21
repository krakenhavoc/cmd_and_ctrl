package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Saddle of the Cavalier — Artifact — Equipment {2}{G} (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Equipped creature gets +3/+3 and can't be blocked by creatures with
//	 power 3 or less.
//	 Equip {3}"
//
// Wrecking Ball Arm's block restriction at 3, with a +3/+3 static. The
// blocker's power is read live as blockers are declared (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6cd939e-6b45-4d47-b251-7a09a149dedd",
		Name:         "Saddle of the Cavalier",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnAttached(), effects.PowerLE(3), "creatures with power 3 or less"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
