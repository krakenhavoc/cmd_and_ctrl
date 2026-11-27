package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Delraich — Creature — Horror {6}{B}, 6/6 (Mercadian Masques, S82 OLD b683,
// #5619):
//
//	"You may sacrifice three black creatures rather than pay this spell's mana
//	 cost.
//	 Trample"
//
// Demon of Death's Gate's alternative cost without the life: three black
// creatures you control, sacrificed as the spell is cast (CR 118.9), all
// validated before any is paid. Trample is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "38f9e766-a2ad-437b-a38c-d6f3ffd261fb",
		Name:            "Delraich",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(3, "three black creatures", 0, effects.OfColor("B"), effects.Creature()),
		},
	})
}
