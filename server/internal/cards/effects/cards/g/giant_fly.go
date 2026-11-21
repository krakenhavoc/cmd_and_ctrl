package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Giant Fly — Creature — Insect {2}{B}, 2/2 (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Flying
//	 Whenever you sacrifice another permanent, this creature gets +1/+0
//	 until end of turn."
//
// Pirate Peddlers' sacrifice trigger (avB18YouSacrificedAnother: one per
// permanent you sacrifice, never the Fly itself) with a pump in place of
// the counter. A Fly that left and came back before the trigger resolves
// is a new object and gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "439ae30b-348d-480c-8d3d-fe105167bdc4",
		Name:            "Giant Fly",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.AvB18YouSacrificedAnother, "Giant Fly — this creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Giant Fly — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
