package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c74f6a83-a280-450e-969d-6b3fa28c8744",
		Name:         "Hedgewitch's Mask",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 1)},
		BlockRules:   []game.BlockRule{CantBeBlockedBy(OnAttached(), PowerGE(4), "creatures with power 4 or greater")},
		Activated:    []ActivatedAbility{EquipAbility("{2}")},
	})
}
