package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "439ae30b-348d-480c-8d3d-fe105167bdc4",
		Name:            "Giant Fly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, avB18YouSacrificedAnother, "Giant Fly — this creature gets +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if sourceIsNewObject(g, item) || !onBattlefield(g, ctx.Source()) {
						return nil
					}
					return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Giant Fly — +1/+0"}.Apply(ctx)
				}),
		},
	})
}
