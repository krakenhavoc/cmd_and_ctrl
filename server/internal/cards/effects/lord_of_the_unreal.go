package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	illusions := TribeFilter{Tribes: []string{"Illusion"}, YoursOnly: true}
	Register(Spec{
		OracleID:     "cc60181a-5eeb-4d66-9876-3194bac1e7a7",
		Name:         "Lord of the Unreal",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(illusions, 1, 1),
			TribalKeywordGrant(illusions, "hexproof"),
		},
	})
}
