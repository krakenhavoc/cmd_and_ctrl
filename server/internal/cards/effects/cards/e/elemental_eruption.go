package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elemental Eruption — Sorcery {4}{R}{R}:
//
//	"Create a 4/4 red Dragon Elemental creature token with flying and
//	 prowess.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn.)"
//
// Batch 53 (#460) listed it as shipping without the token's prowess.
// Since #706 prowess is a keyword the engine turns into a trigger, so
// the token row's Keywords carry it and nothing else is needed.
//
// Two rules meet here and the card is a clean test of both. Each
// storm copy is a separate resolution that makes its own Dragon
// Elemental (Empty the Warrens has the same shape). And the copies are
// NOT cast (CR 707.10), so a Dragon Elemental already on the
// battlefield gets +1/+1 for the Eruption itself but nothing for its
// copies — prowess reads EventCast, which a copy never emits.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8629fdef-dbbe-4a97-a7e1-6b1646f4ca0b",
		Name:         "Elemental Eruption",
		Purpose:      game.Purpose{Tokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Storm()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon Elemental with flying and prowess"), N: 1}.Apply(ctx)
		},
	})
}
