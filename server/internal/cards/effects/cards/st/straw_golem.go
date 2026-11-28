package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Straw Golem — Artifact Creature — Golem {1}, 2/3 (Weatherlight):
//
//	"When an opponent casts a creature spell, sacrifice this creature."
//
// Triggers on the cast, so the Golem is gone before the spell resolves; a
// creature spell its own controller casts does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e48981b-8971-4b7c-ae05-29c7e4091a05",
		Name:         "Straw Golem",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Creature()),
				"Straw Golem — sacrifice this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
