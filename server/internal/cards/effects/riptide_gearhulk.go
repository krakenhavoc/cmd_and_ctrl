package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := WhenThisEnters("Riptide Gearhulk — for each opponent, put up to one target nonland permanent that player controls into its owner's library third from the top",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := b22TuckThirdFromTop(g, t.ID); err != nil {
					return err
				}
			}
			return nil
		})
	trigger.Targets = perOpponentClauses(game.MaxPlayers-1, "nonland permanent", Nonland())
	Register(Spec{
		OracleID:        "f3838ff1-5403-4f3f-8531-36029c13fcfc",
		Name:            "Riptide Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "prowess"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
