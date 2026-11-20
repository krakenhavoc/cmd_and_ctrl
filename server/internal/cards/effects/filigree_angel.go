package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Filigree Angel — Artifact Creature — Angel {5}{W}{W}{U}, 4/4 (Alara
// Reborn, #5586):
//
//	"Flying
//	 When this creature enters, you gain 3 life for each artifact you
//	 control."
//
// The artifacts are counted as the trigger resolves (CR 608.2h), after the
// layers, so the Angel counts itself while it is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a261c42-825d-4965-8ea7-80d50c43474e",
		Name:            "Filigree Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Filigree Angel — you gain 3 life for each artifact you control",
				func(g *game.Game, item *game.StackItem) error {
					n := rfCreatureAArtifactsYouControl(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return GainLife{Player: item.Controller, Amount: 3 * n}.Apply(NewContext(g, item))
				}),
		},
	})
}
