package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Demonmail Hauberk — Artifact — Equipment {4} (Innistrad Remastered, #5948):
//
//	"Equipped creature gets +4/+2.
//	 Equip—Sacrifice a creature."
//
// The equip's cost is a sacrifice of any creature (EquipPayingAbility, as
// Bloodthorn Flail pays a branch): chosen as the ability is announced, so the
// creature is gone before the equip resolves, and it may be the creature the
// Hauberk would have equipped, in which case the equip has no target left to
// attach to and does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dac85722-d741-4584-8b98-8ac220958a38",
		Name:         "Demonmail Hauberk",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(4, 2)},
		Activated: []ActivatedAbility{
			EquipPayingAbility("Equip—Sacrifice a creature.", SacrificeN(1, "a creature", Creature())),
		},
	})
}
