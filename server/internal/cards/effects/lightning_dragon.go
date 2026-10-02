package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Dragon — Creature — Dragon, {2}{R}{R}, 4/4:
//
//	"Flying
//	 Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {R}: This creature gets +1/+0 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be0fc708-b205-4307-85f3-f8fd1efd4cf6",
		Name:            "Lightning Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{R}: This creature gets +1/+0 until end of turn.",
			Cost:  ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Lightning Dragon — +1/+0"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Lightning Dragon", "{2}{R}{R}")},
	})
}
