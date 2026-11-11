package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krosan Cloudscraper — Creature — Beast Mutant {7}{G}{G}{G}, 13/13:
//
//	"At the beginning of your upkeep, sacrifice this creature unless you pay {G}{G}.
//	 Morph {7}{G}{G}"
//
// The upkeep payment is Haazda Shield Mate's shape: pay {G}{G} or
// sacrifice it. A face-down Cloudscraper has no text (CR 708.2a), so it
// asks for nothing until it is turned face up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a0ff742b-f709-43ac-84d2-dc00c723bad6",
		Name:             "Krosan Cloudscraper",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{7}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Krosan Cloudscraper — sacrifice unless you pay {G}{G}", func(g *game.Game, item *game.StackItem) error {
				return UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{G}{G}",
					Question: "Krosan Cloudscraper — pay {G}{G} or sacrifice Krosan Cloudscraper?",
					Else:     upkeepSacrificeSource,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
