package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trove Tracker — Creature — Human Pirate {2}{U}, 2/2 (Commander Legends,
// #5507):
//
//	"When this creature dies, draw a card.
//	 Encore {5}{U}{U}"
//
// The dies trigger draws for its controller; encore is the shared
// constructor. Each encore copy that dies draws a card too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f27972da-ad2f-4df5-82d3-c909cce8cc91",
		Name:         "Trove Tracker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Trove Tracker — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{Encore("{5}{U}{U}")},
	})
}
