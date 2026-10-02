package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Moon — Enchantment {2}{R}:
//
//	"Nonbasic lands are Mountains."
//
// ADR 0109 owner decision 4 (#1881): Magus of the Moon's static, the
// static form of CR 305.7 (effects.SetsBasicLandType) over every nonbasic
// land, everyone's. In layer 4 each one's land types are replaced by Mountain and
// every other subtype stays (CR 205.1a), it loses the abilities its rules
// text gives it and keeps any another effect granted it, and it taps for
// {R} (CR 305.6). Basic lands are untouched: the test is the supertype (CR
// 205.4c), so a Snow-Covered Forest stays a Forest. Against Urborg, Tomb of
// Yawgmoth the two settle by timestamp in the same layer (CR 613.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94fac5fe-97d5-4c12-a80c-8efff9d853ae",
		Name:         "Blood Moon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SetsBasicLandType(nonbasicLand, nil, []string{"Mountain"}),
		},
	})
}
