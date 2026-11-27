package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rhox — Creature — Rhino Beast {4}{G}{G}, 5/5 (Eighth Edition, S82 OLD batch 612,
// #5607):
//
//	"You may have this creature assign its combat damage as though it weren't
//	 blocked.
//	 {2}{G}: Regenerate this creature."
//
// Pride of Lions' static (a layer 6 ability on the Rhox alone, an option over the
// blocked assignment of CR 510.1c, lost with its abilities) and Drudge
// Skeletons' regeneration row at {2}{G} (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6f4c684-2915-4a58-9c4b-487387f7cf73",
		Name:         "Rhox",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have this creature assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
		Activated: []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{2}{G}")},
	})
}
