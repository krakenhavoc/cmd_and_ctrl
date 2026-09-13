package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauding Blight-Priest — Creature — Vampire Cleric {2}{B}, 3/2
// (EDHREC rank 1113):
//
//	"Whenever you gain life, each opponent loses 1 life."
//
// The lifegain deck's table-wide drain. Sanguine Bond's trigger with
// a fixed 1 and no target: every EventChangeLife with a positive
// delta on the controller — lifelink included — fires it once, and
// "gain 5 life" is one event, so one trigger, as printed. Life LOSS,
// not damage, so no prevention or doubling applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "814b87fe-2a75-4ff2-8637-7e69e3fb285b",
		Name:         "Marauding Blight-Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b10YouGainedLife(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Marauding Blight-Priest — each opponent loses 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return eachOpponentLosesLife(g, item, 1)
					})
			},
		}},
	})
}
