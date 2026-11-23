package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedgewitch's Mask — Artifact — Equipment {W} (Innistrad: Midnight Hunt):
//
//	"Equipped creature gets +1/+1.
//	 Equipped creature can't be blocked by creatures with power 4 or greater.
//	 Equip {2}"
//
// The pump is PumpAttached and the evasion is a block rule on the equipped
// creature (Kithkin Armor's shape), read when blockers are declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c74f6a83-a280-450e-969d-6b3fa28c8744",
		Name:         "Hedgewitch's Mask",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 1)},
		BlockRules:   []game.BlockRule{effects.CantBeBlockedBy(effects.OnAttached(), effects.PowerGE(4), "creatures with power 4 or greater")},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
