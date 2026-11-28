package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rejuvenate — Sorcery {3}{G} (Urza's Saga, S82 OLD b722):
//
//	You gain 6 life.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//
// Six life for the caster, with cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35069cdb-e5bd-4224-a1a8-b329054e003b",
		Name:         "Rejuvenate",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Player: ctx.Controller(), Amount: 6}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
