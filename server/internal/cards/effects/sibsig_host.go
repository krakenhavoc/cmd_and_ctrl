package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sibsig Host — Creature — Zombie {4}{B}, 2/6 (Fate Reforged,
// #5552):
//
//	"When this creature enters, each player mills three cards."
//
// Eye Collector's each-player mill on an entry trigger: every player, in
// turn order from the active player, mills three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11776f73-f504-4f19-b5b6-2f58279f8b15",
		Name:         "Sibsig Host",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sibsig Host — each player mills three cards",
				func(g *game.Game, item *game.StackItem) error { return b02EachPlayerMills(g, item, 3) }),
		},
	})
}
