package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kiln Fiend — Creature — Elemental Beast {1}{R}, 1/2:
//
//	"Whenever you cast an instant or sorcery spell, this creature gets
//	 +3/+0 until end of turn."
//
// Kinsbaile Aspirant's self-pump on Archmage of Runes' cast watcher.
//
// No simplification.
func init() {
	const label = "Kiln Fiend — +3/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "eac8c196-8477-4b79-9875-21afa1e61708",
		Name:         "Kiln Fiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 3, Label: label}.Apply(ctx)
			}),
		},
	})
}
