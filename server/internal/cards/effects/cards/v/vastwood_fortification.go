package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vastwood Fortification // Vastwood Thicket — Modal DFC. Front: Instant
// {G}: "Put a +1/+1 counter on target creature." Back: Land
// (mdfc_lands.go).
//
// One counter through the counter pipeline, so Hardened Scales and
// Doubling Season see it. A target that is no longer legal means the spell
// does nothing. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce148a0c-6c63-49d5-a156-99efae4e367a",
		Name:         "Vastwood Fortification",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutPlusOneCounterOnTheTarget(ctx)
		},
	})
}
