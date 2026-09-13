package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Castle Vantress — Land (EDHREC rank 1008):
//
//	"This land enters tapped unless you control an Island.
//	 {T}: Add {U}.
//	 {2}{U}{U}, {T}: Scry 2."
//
// Blue's Castle: Castle Locthwain's shape with a scry where the
// Locthwain draws. The scry is the Scry primitive, which queues the
// look-and-reorder prompt and moves nothing until the controller
// answers; "unless you control an Island" reads the post-layer land
// type, so an Urborg-style Island counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdf41cf4-4e77-453d-be5b-0abbbd358934",
		Name:         "Castle Vantress",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("island")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}{U}{U}, {T}: Scry 2.",
			Cost:  Plus(ManaCost("{2}{U}{U}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
