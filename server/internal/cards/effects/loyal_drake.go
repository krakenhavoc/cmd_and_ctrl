package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Drake — Creature — Drake {2}{U}, 2/2 (Commander 2018, #5526):
//
//	"Flying
//	 Lieutenant — At the beginning of combat on your turn, if you control
//	 your commander, draw a card."
//
// Lieutenant is an ability word, so the condition is an intervening if (CR
// 603.4), Loyal Apprentice's: checked as your beginning of combat starts and
// again as the trigger resolves (b18ControlsYourCommander, a commander you
// own and control).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "daebf011-ec85-4cbd-9445-02ac59675c58",
		Name:            "Loyal Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, c18LieutenantCombat, "Loyal Drake — draw a card",
				func(g *game.Game, item *game.StackItem) error {
					if !b18ControlsYourCommander(g, item.Controller) {
						return nil
					}
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}

// c18LieutenantCombat is the lieutenant condition of the Commander 2018
// Loyal cycle: your beginning of combat, while you control your commander.
// Each resolution checks the commander again.
func c18LieutenantCombat(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
		b18ControlsYourCommander(g, source.Controller)
}
