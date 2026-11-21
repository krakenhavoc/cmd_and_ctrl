package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "2d513c3e-bd27-479e-af32-7f66d96d0de0",
		Name:            "Yeva, Nature's Herald",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{CreatureOnly: true, ColorsAny: "G"},
				"You may cast green creature spells as though they had flash."),
		},
	})
}
