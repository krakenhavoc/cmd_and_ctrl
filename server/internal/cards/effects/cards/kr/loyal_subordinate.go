package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "76f405a1-242a-47e6-9aa6-261e49a85b45",
		Name:            "Loyal Subordinate",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, effects.C18LieutenantCombat, "Loyal Subordinate — each opponent loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B18ControlsYourCommander(g, item.Controller) {
						return nil
					}
					return effects.EachOpponentLosesLife{N: 3}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
