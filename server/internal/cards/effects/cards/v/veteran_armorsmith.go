package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Armorsmith — Creature — Human Soldier {W}{W}, 2/3 (Magic 2010, #5585):
//
//	"Other Soldier creatures you control get +0/+1."
//
// Yotian Tactician's lord with +0/+1: the other Soldiers you control, and
// not an opponent's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcb34e92-f991-401c-a494-c2919083bf9e",
		Name:         "Veteran Armorsmith",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, 0, 1),
		},
	})
}
