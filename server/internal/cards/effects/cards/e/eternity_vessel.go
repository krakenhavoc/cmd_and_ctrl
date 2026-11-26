package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eternity Vessel — Artifact {6} (Zendikar, S82 OLD b496):
//
//	"This artifact enters with X charge counters on it, where X is your life total.
//	 Landfall — Whenever a land you control enters, you may have your life total
//	 become the number of charge counters on this artifact."
//
// Academy Elite's counted entry replacement over the controller's life total as
// the Vessel enters; the landfall is optional and sets the life total by gaining or
// losing the difference (CR 119.5, B31LifeBecomes), so lifegain and life-loss
// triggers see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8fa1319-6623-4999-acb5-a5e15e6d868c",
		Name:         "Eternity Vessel",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted("charge",
				func(g *game.Game, src *game.Card) int {
					if p := g.PlayerByIDForEffect(src.Controller); p != nil && p.Life > 0 {
						return p.Life
					}
					return 0
				},
				"Eternity Vessel enters with X charge counters on it, where X is your life total"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Landfall("Eternity Vessel — your life total becomes the number of charge counters on it", effects.Old496EternityVesselBecomes),
				"Eternity Vessel — have your life total become the number of charge counters on it?"),
		},
	})
}
