package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thunderous Snapper — Creature — Turtle Hydra {G/U}{G/U}{G/U}{G/U}, 4/4 (Throne of Eldraine, #5516):
//
//	"Whenever you cast a spell with mana value 5 or greater, draw a card."
//
// Skybeast Tracker's cast trigger, drawing instead of making Food.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f826ac35-7dde-414c-94ec-0bdfe04327b4",
		Name:         "Thunderous Snapper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(ManaValueGE(5), "Thunderous Snapper — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
