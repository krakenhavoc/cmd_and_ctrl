package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Riptide Gearhulk — Artifact Creature — Construct {1}{W}{W}{U}{U}, 2/5
// (Aetherdrift, #4464):
//
//	"Double strike
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, for each opponent, put up to one target
//	 nonland permanent that player controls into its owner's library third
//	 from the top."
//
// "For each opponent, up to one target … that player controls" is Celebrate
// the Mountain-king's per-opponent clauses (perOpponentClauses): one
// "up to one" clause per opponent, each bound to that opponent, so the
// CR 608.2b re-check asks whether THAT player still controls the pick. Each
// pick still legal goes into its owner's library third from the top, or on
// the bottom of a library of fewer than two cards (God-Eternal Oketra's
// positioned landing), one after another in printed order. Double strike
// and prowess are the engine's keywords (CR 702.4, 702.108).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Riptide Gearhulk — for each opponent, put up to one target nonland permanent that player controls into its owner's library third from the top",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := effects.B22TuckThirdFromTop(g, t.ID); err != nil {
					return err
				}
			}
			return nil
		})
	trigger.Targets = effects.PerOpponentClauses(game.MaxPlayers-1, "nonland permanent", effects.Nonland())
	effects.Register(effects.Spec{
		OracleID:        "f3838ff1-5403-4f3f-8531-36029c13fcfc",
		Name:            "Riptide Gearhulk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", "prowess"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
