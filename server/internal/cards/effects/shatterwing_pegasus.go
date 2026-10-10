package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shatterwing Pegasus — Creature — Pegasus {2}{W}, 2/3:
//
//	"Flying
//	 {4}{W}: Creatures you control get +1/+1 until end of turn."
//
// The pump covers the creatures its controller has when the ability
// resolves, not ones that arrive later (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7d6641e7-4d14-4a06-b588-0c752af529bd",
		Name:            "Shatterwing Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{4}{W}: Creatures you control get +1/+1 until end of turn.",
			Cost:  ManaCost("{4}{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Shatterwing Pegasus — creatures you control get +1/+1",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
