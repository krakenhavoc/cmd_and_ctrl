package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lord of the Unreal — Creature — Human Wizard {U}{U}, 2/2:
//
//	"Illusion creatures you control get +1/+1 and have hexproof."
//
// A tribal anthem and a tribal keyword grant over the effective subtype
// (a changeling counts). It is not an Illusion itself, so it pumps
// nothing but other Illusions you control.
//
// No simplification.
func init() {
	illusions := effects.TribeFilter{Tribes: []string{"Illusion"}, YoursOnly: true}
	effects.Register(effects.Spec{
		OracleID:     "cc60181a-5eeb-4d66-9876-3194bac1e7a7",
		Name:         "Lord of the Unreal",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(illusions, 1, 1),
			effects.TribalKeywordGrant(illusions, "hexproof"),
		},
	})
}
