package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Widespread Panic — Enchantment {2}{R} (Commander 2013, S82 OLD b386):
//
//	"Whenever a spell or ability causes its controller to shuffle their
//	 library, that player puts a card from their hand on top of their
//	 library."
//
// Watches game.EventLibraryShuffled caused by a spell or ability (a search's
// shuffle included; the sandbox's manual shuffle and the opening-hand and
// mulligan shuffles are no spell or ability). "That player" is the shuffled
// library's owner, the event's actor, who picks the card (Brainstorm's put
// back, one card). An empty hand puts nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "853a3c2b-3d37-453a-8a77-4d90bd3a1cb7",
		Name:         "Widespread Panic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLibraryShuffled, effects.ASpellOrAbilityShufflesALibrary,
				"Widespread Panic — that player puts a card from their hand on top of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.PutFromHandOnTopInAnyOrder{
						Player: ctx.Trigger().Event.Actor,
						N:      1,
						Label:  "Widespread Panic — put a card from your hand on top of your library",
					}.Apply(ctx)
				}),
		},
	})
}
