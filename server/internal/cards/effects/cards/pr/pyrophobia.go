package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrophobia — Sorcery {1}{R} (Foundations Jumpstart):
//
//	"Pyrophobia deals 3 damage to target creature. Cowards can't block
//	 this turn."
//
// Damage to the target, then one live "Cowards can't block" record over
// every creature (CantBlockSubtypesThisTurn). The rule does not need the
// target: the damage is skipped if the target is gone, but the sentence
// that follows still applies (the spell resolves while any target is
// legal, CR 608.2b). Creature types are read live, so a changeling is a
// Coward.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa70903d-3c8b-49f7-8817-f708b5ba250f",
		Name:         "Pyrophobia",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CantBlockSubtypesThisTurn{Subtypes: ctx.TextSubtypes("Coward"), Text: "Cowards", Label: "Pyrophobia — Cowards can't block"}.Apply(ctx)
		},
	})
}
