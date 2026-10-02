package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Basalt Gargoyle — Creature — Gargoyle, {2}{R}, 3/2:
//
//	"Flying
//	 Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 {R}: This creature gets +0/+1 until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83a38c8d-e2f6-404b-83be-9cfce037d6e2",
		Name:            "Basalt Gargoyle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{R}: This creature gets +0/+1 until end of turn.",
			Cost:  ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Toughness: 1, Label: "Basalt Gargoyle — +0/+1"}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{Echo("Basalt Gargoyle", "{2}{R}")},
	})
}
