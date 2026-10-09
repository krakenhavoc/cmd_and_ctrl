package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ob Nixilis, the Ascended — Legendary Creature — Angel {5}{W}{W}, 4/4:
//
//	"Flying
//	 When Ob Nixilis enters, destroy all tapped creatures your
//	 opponents control. You gain 1 life for each creature destroyed this
//	 way.
//	 At the beginning of each end step, if you gained life this turn,
//	 create a 4/4 white Angel creature token with flying."
//
// The life is the number of creatures that were actually destroyed (an
// indestructible creature is not counted). The end-step trigger has an
// intervening "if" (CR 603.4), checked when each end step begins and
// again as the trigger resolves, off the controller's life gained this
// turn — so lifelink on an opponent's turn counts at that turn's end.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a49b518-61c8-4dea-8008-c5edf8384581",
		Name:            "Ob Nixilis, the Ascended",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ob Nixilis, the Ascended — destroy all tapped creatures your opponents control", func(g *game.Game, item *game.StackItem) error {
				controller := item.Controller
				return DestroyAllMatching{
					Match: And(Creature(), OpponentControls(), rfCreatureDTapped()),
					Then: func(ctx *Context, _ []game.Card, destroyed int) error {
						return GainLife{Player: controller, Amount: destroyed}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 1)
			}, "Ob Nixilis, the Ascended — create a 4/4 Angel with flying", b32AngelIfYouGainedLifeThisTurnAtLeast(1)),
		},
	})
}
