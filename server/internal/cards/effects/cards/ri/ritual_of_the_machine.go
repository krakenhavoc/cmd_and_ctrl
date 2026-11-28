package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ritual of the Machine — Sorcery {2}{B}{B} (Alliances, S82 OLD batch 817):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Gain control of target nonartifact, nonblack creature."
//
// The theft has no duration, so it lasts until the game ends or another
// effect takes the creature (CR 611.2a). The creature is sacrificed with
// the spell already on the stack; if it was the target, the spell fizzles.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bccd8298-e056-4b91-b090-408c75467b92",
		Name:           "Ritual of the Machine",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetCreature("target nonartifact, nonblack creature", effects.Not(effects.Artifact()), effects.NonBlack()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GainControl{Target: id, Label: "Ritual of the Machine — gain control"}.Apply(ctx)
		},
	})
}
