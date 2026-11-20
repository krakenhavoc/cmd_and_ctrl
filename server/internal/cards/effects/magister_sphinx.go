package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magister Sphinx — Artifact Creature — Sphinx {4}{W}{U}{B}, 5/5
// (Conflux, #5587):
//
//	"Flying
//	 When this creature enters, target player's life total becomes 10."
//
// Sorin Markov's "becomes 10" on an enters trigger: a life change of
// whatever size closes the gap (CR 119.5), so a player above 10 loses
// life and one below gains it. Any player may be the target, you
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1d86ae29-4705-4bbc-8922-9b4e031df5b1",
		Name:            "Magister Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Magister Sphinx — target player's life total becomes 10",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					p := targetedPlayerOf(ctx)
					if p == uuid.Nil {
						return nil
					}
					return b31LifeBecomes(ctx, p, 10)
				}),
				TargetPlayer("target player")),
		},
	})
}
