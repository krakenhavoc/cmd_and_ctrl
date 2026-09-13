package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedron Crab — Creature — Crab {U}, 0/2 (EDHREC rank 1422):
//
//	"Landfall — Whenever a land you control enters, target player
//	 mills three cards."
//
// The mill deck's one-drop. A targeted landfall trigger: the
// condition is Tireless Provisioner's (any land entering under the
// controller's control — played, fetched or returned), the target is
// a real target clause the engine prompts for when the trigger goes
// on the stack, and the mill is the ordinary primitive, so an empty
// library is the SBA's problem and not the Crab's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7216f974-3c84-40ef-b904-82019900c204",
		Name:         "Hedron Crab",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13LandYouControlEntered(ev, source, g)
			},
			Targets: TargetPlayer("target player"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Hedron Crab — target player mills three cards",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						return MillCards{Player: item.Targets[0].ID, N: 3}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
