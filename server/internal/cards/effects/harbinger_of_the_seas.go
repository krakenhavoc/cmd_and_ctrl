package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harbinger of the Seas — Creature — Merfolk Wizard {1}{U}{U}, 2/2:
//
//	"Nonbasic lands are Islands."
//
// ADR 0109 owner decision 4 (#1881): Magus of the Moon's static, the
// static form of CR 305.7 (effects.SetsBasicLandType) over every nonbasic
// land, everyone's. In layer 4 each one's land types are replaced by Island and
// every other subtype stays (CR 205.1a), it loses the abilities its rules
// text gives it and keeps any another effect granted it, and it taps for
// {U} (CR 305.6). Basic lands are untouched: the test is the supertype (CR
// 205.4c), so a Snow-Covered Forest stays a Forest. Against Urborg, Tomb of
// Yawgmoth the two settle by timestamp in the same layer (CR 613.7).
//
// The Islands version on a creature, as Magus of the Moon is Blood Moon's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84910453-e5dc-4bff-8554-86a5f4ab2746",
		Name:         "Harbinger of the Seas",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SetsBasicLandType(nonbasicLand, nil, []string{"Island"}),
		},
	})
}
