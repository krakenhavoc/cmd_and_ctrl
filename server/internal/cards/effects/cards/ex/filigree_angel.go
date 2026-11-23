package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "4a261c42-825d-4965-8ea7-80d50c43474e",
		Name:            "Filigree Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Filigree Angel — you gain 3 life for each artifact you control",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.RfCreatureAArtifactsYouControl(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.GainLife{Player: item.Controller, Amount: 3 * n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
