package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yeva, Nature's Herald — Legendary Creature — Elf Shaman {2}{G}{G}, 4/4:
//
//	"Flash
//	 You may cast green creature spells as though they had flash."
//
// Vivien, Champion of the Wilds' creature-only flash timing, narrowed to
// green by ColorsAny (a multicoloured green creature qualifies, CR 105.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d513c3e-bd27-479e-af32-7f66d96d0de0",
		Name:            "Yeva, Nature's Herald",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true, ColorsAny: "G"},
				"You may cast green creature spells as though they had flash."),
		},
	})
}
