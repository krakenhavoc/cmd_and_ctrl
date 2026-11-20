package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drifter il-Dal — Creature — Human Wizard {U}, 2/1 (Time Spiral, #5596):
//
//	"Shadow
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {U}."
//
// Krosan Cloudscraper's upkeep tax at {U}: a decline, or a "yes" that
// can't be funded, sacrifices it (upkeepSacrificeSource).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbe8ad46-d862-4ff6-bf17-238fb8341075",
		Name:            "Drifter il-Dal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Drifter il-Dal — sacrifice unless you pay {U}", func(g *game.Game, item *game.StackItem) error {
				return UpkeepPayUnless{
					Chooser:  item.Controller,
					Cost:     "{U}",
					Question: "Drifter il-Dal — pay {U} or sacrifice Drifter il-Dal?",
					Else:     upkeepSacrificeSource,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
