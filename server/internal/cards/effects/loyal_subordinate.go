package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Subordinate — Creature — Zombie {2}{B}, 3/1 (Commander 2018,
// #5526):
//
//	"Menace
//	 Lieutenant — At the beginning of combat on your turn, if you control
//	 your commander, each opponent loses 3 life."
//
// Loyal Drake's lieutenant trigger (c18LieutenantCombat, an intervening if
// checked again as it resolves) with Scheming Aspirant's life loss.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "76f405a1-242a-47e6-9aa6-261e49a85b45",
		Name:            "Loyal Subordinate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, c18LieutenantCombat, "Loyal Subordinate — each opponent loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					if !b18ControlsYourCommander(g, item.Controller) {
						return nil
					}
					return EachOpponentLosesLife{N: 3}.Apply(NewContext(g, item))
				}),
		},
	})
}
