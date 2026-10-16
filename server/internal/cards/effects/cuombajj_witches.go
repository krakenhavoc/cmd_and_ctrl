package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cuombajj Witches — Creature — Human Wizard {B}{B}, 1/3:
//
//	"{T}: This creature deals 1 damage to any target and 1 damage to any
//	 target of an opponent's choice."
//
// Two instances of "target", so two clauses (CR 601.2c): the first is
// the controller's, the second an opponent's choice (#2055, ADR 0019's
// 2026-10-16 amendment). The rulings: the controller chooses which
// opponent chooses the second target, and it doesn't have to be the
// opponent who controls (or is) the first; the controller still controls
// the ability, so an opponent can't target their own hexproof creature.
// The two may name the same object (two instances of "target"), and each
// is checked on its own as the ability resolves (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "638eeb16-9e0e-4cc6-b97e-8ff0df81ca58",
		Name:         "Cuombajj Witches",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: This creature deals 1 damage to any target and 1 damage to any target of an opponent's choice.",
			Cost:    TapCost(),
			Targets: Clauses(TargetAny(), OfAnOpponentsChoice(cuombajjAnyTarget())),
			Purpose: ForTargets(DamageToTarget(0, 1), DamageToTarget(1, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return g.DamageInstanceForEffect(func() error {
					for slot := 0; slot < 2; slot++ {
						t, ok := ctx.ClauseTarget(slot)
						if !ok {
							continue
						}
						if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}

// cuombajjAnyTarget is "any target of an opponent's choice": any target
// (CR 115.4), worded for the opponent's prompt.
func cuombajjAnyTarget() *game.TargetSpec {
	s := TargetAny()
	s.Label = "any target of an opponent's choice"
	return s
}
