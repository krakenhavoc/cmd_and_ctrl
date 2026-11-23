package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jadelight Spelunker — Creature — Merfolk Scout {X}{G}, 1/1 (Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, it explores X times."
//
// X is the cast's X, stamped on the entry trigger by the engine (XFromCast,
// as Champions from Beyond's is). The explores run one after another through
// the engine's explore-N seam (game.ExploreTimesForEffect): each one's
// graveyard question is answered before the next reveals, so a nonland card
// left on top is revealed again (the ruling). Entering without being cast,
// or cast with X=0, it does not explore at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a72d6cd-b4a7-4c42-90ba-a14f1733d00e",
		Name:         "Jadelight Spelunker",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       "Jadelight Spelunker — explores X times",
			XFromCast: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ExploresTimes{Explorer: item.SourceCardID, N: ctx.X()}.Apply(ctx)
			},
		}},
	})
}
