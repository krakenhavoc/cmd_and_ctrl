package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jumbo Cactuar — Creature — Plant {5}{G}{G}, 1/7:
//
//	"10,000 Needles — Whenever this creature attacks, it gets +9999/+0
//	 until end of turn."
//
// One attack trigger whose effect is a plain until-end-of-turn boost of
// its own source (CR 611.2a).
func init() {
	Register(Spec{
		OracleID:     "e0bdfbb4-3060-4492-bd9b-0c47e719400a",
		Name:         "Jumbo Cactuar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("10,000 Needles — it gets +9999/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{Target: ctx.Source(), Power: 9999, Label: "10,000 Needles — +9999/+0"}.Apply(ctx)
				}),
		},
	})
}
