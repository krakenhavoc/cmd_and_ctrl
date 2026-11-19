package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "eac8c196-8477-4b79-9875-21afa1e61708",
		Name:         "Kiln Fiend",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 3, Label: label}.Apply(ctx)
			}),
		},
	})
}
