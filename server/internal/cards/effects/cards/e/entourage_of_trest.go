package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entourage of Trest — Creature — Elf Soldier {4}{G}, 4/4:
//
//	"When this creature enters, you become the monarch.
//	 This creature can block an additional creature each combat as
//	 long as you're the monarch."
//
// Skipped by the #1720 builder for want of the monarch seam; #1722 is
// that seam. The block line is CanBlockAdditional (#1706) on itself,
// gated on YoureTheMonarch. The gate is a LAYER INPUT — the crown can
// change hands with no permanent moving — so the layer pass has to be
// told when it does: layerVersionBump bumps on EventMonarchChanged,
// and a steal in combat damage takes the extra block away before the
// next combat's declare-blockers step reads it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27b0bd3f-2c11-474a-8512-fe035a514905",
		Name:         "Entourage of Trest",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Entourage of Trest"),
		},
		Static: []game.StaticAbility{effects.CanBlockAdditional(selfWhileYoureTheMonarch, 1)},
	})
}

// selfWhileYoureTheMonarch is "this creature … as long as you're the
// monarch".
func selfWhileYoureTheMonarch(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID && effects.YoureTheMonarch(g, source.Controller)
}
