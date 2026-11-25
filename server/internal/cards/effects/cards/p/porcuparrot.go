package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Porcuparrot — Creature — Bird Beast {3}{R}, 3/4 (Ikoria: Lair of Behemoths,
// #5512):
//
//	"Mutate {2}{R} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 {T}: This creature deals X damage to any target, where X is the number of
//	 times this creature has mutated."
//
// Mutate is the engine's (ADR 0223). X is the merged permanent's count of
// mutations (S82OldB111TimesMutated), read as the ability resolves, so a
// Porcuparrot that never mutated deals 0 damage (CR 120.8). Its abilities work
// from under another creature, and the {T} is the merged permanent's own, so
// summoning sickness is the permanent's (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "e1efd99f-d918-4cd1-b85f-b90b106c3c88",
		Name:             "Porcuparrot",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{R}")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: This creature deals X damage to any target, where X is the number of times this creature has mutated.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.S82OldB111TimesMutated(g, item.SourceCardID)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: x}.Apply(ctx)
			},
		}},
	})
}
