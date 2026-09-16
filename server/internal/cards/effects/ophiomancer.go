package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ophiomancer — Creature — Human Shaman {2}{B}, 2/2 (EDHREC rank 645):
//
//	"At the beginning of each upkeep, if you control no Snakes,
//	 create a 1/1 black Snake creature token with deathtouch."
//
// A deathtouch blocker every turn, and a sacrifice-fodder engine
// with any outlet: sacrifice the Snake on your turn, get a new one on
// each opponent's upkeep. "EACH upkeep" — no controller gate on the
// event.
//
// The intervening-if (CR 603.4) is checked BOTH times: in AppliesTo,
// so no trigger goes on the stack while a Snake is out, and again at
// resolution, so a Snake that arrives in response (an opponent's
// Ophiomancer trigger resolving first, a token copy) stops the
// second one. Most intervening-if cards in the catalog check only at
// trigger time; this one can afford the re-check because the condition
// is a board read with no target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55eca80c-dcd8-4c2f-aa0f-fb0aec7b80f7",
		Name:         "Ophiomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return !b05ControlsSubtype(g, source.Controller, "Snake")
			}, "Ophiomancer — create a 1/1 Snake with deathtouch", func(g *game.Game, item *game.StackItem) error {
				if b05ControlsSubtype(g, item.Controller, "Snake") {
					return nil // CR 603.4: the condition is re-checked on resolution
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("1/1 colorless Snake with deathtouch"),
					N:          1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
