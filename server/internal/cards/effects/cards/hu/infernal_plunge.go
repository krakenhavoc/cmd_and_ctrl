package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Plunge — Sorcery {R} (Innistrad, S82 OLD batch 450, #5572):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Add {R}{R}{R}."
//
// Morbid Curiosity's sacrifice cost over Dark Ritual's mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0b67dff0-5bb7-44fe-a72b-d3b1a6c0ee72",
		Name:           "Infernal Plunge",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{Produced: "{R}{R}{R}"}.Apply(ctx)
		},
	})
}
