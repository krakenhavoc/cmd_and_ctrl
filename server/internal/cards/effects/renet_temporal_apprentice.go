package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Renet, Temporal Apprentice — Legendary Creature — Human Wizard
// {3}{U}{U}, 4/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flash
//	 When Renet enters, return each other nonland permanent that entered
//	 this turn to its owner's hand."
//
// The enters trigger is a mass bounce (BounceAllMatching, one
// simultaneous event) over the per-turn entry tally, read as the trigger
// resolves: every nonland permanent other than this Renet that entered
// the battlefield this turn, whoever controls it. "Other" is this
// object (CR 109.1), so a second Renet that entered this turn goes back.
// A permanent that has been on the battlefield since an earlier turn
// stays, and so does a land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6c754808-86a6-43d3-8aeb-96ec19f7396c",
		Name:            "Renet, Temporal Apprentice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Renet, Temporal Apprentice — return each other nonland permanent that entered this turn to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					return BounceAllMatching{
						Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.InstanceID != self && !c.IsLand() && g.EnteredThisTurn(c.InstanceID)
						},
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
