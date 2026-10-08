package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avishkar Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {3}, {T}, Discard a card: Draw a card."
//
// ADR 0138 (#2122). The rummage is an ordinary activated ability with a
// discard cost (paid at activation, CR 602.2b), activatable only while
// its controller has max speed (CR 702.178a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2c35551-1ba5-4424-baf9-821b49bbcc8c",
		Name:            "Avishkar Raceway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label: "Max speed — {3}, {T}, Discard a card: Draw a card.",
				Cost:  Plus(ManaCost("{3}"), TapCost(), DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			}),
		},
	})
}
