package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Culling the Weak — Instant {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Add {B}{B}{B}{B}."
//
// Burnt Offering's cost with Dark Ritual's resolution: the sacrifice is
// paid at announce (CR 601.2h), so the spell can't be cast without a
// creature to give up, and the four black mana land in the pool as it
// resolves.
//
// No simplification. Mass Casualty deck request (#3245).
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6385fd0d-256d-4400-9319-afa8170024db",
		Name:           "Culling the Weak",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.AddMana{Produced: "{B}{B}{B}{B}"}.Apply(ctx)
		},
	})
}
