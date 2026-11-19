package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fa70903d-3c8b-49f7-8817-f708b5ba250f",
		Name:         "Pyrophobia",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return CantBlockSubtypesThisTurn{Subtypes: ctx.TextSubtypes("Coward"), Text: "Cowards", Label: "Pyrophobia — Cowards can't block"}.Apply(ctx)
		},
	})
}
