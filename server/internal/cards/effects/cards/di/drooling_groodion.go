package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drooling Groodion — Creature — Beast {3}{B}{B}{G}, 4/3 (Ravnica: City of
// Guilds, S82 OLD batch 575, #5600):
//
//	"{2}{B}{G}, Sacrifice a creature: Target creature gets +2/+2 until end of
//	 turn. Another target creature gets -2/-2 until end of turn."
//
// Two clauses, the second a different object from the first (Distinct), both
// chosen as the ability goes on the stack, before the sacrifice is paid, so the
// sacrificed creature may be one of them. Each is re-checked as the ability
// resolves (CR 608.2b) and one that is no longer legal gets nothing while the
// other still does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b027a80b-30cf-4e7b-8902-efd584e0edf7",
		Name:         "Drooling Groodion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}{G}, Sacrifice a creature: Target creature gets +2/+2 until end of turn. Another target creature gets -2/-2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2), effects.HarmUntilEOT(1, -2, -2)), game.AnswerRemove|game.AnswerPump|game.AnswerSacOutlet),
			Cost:    effects.Plus(effects.ManaCost("{2}{B}{G}"), effects.SacrificeN(1, "a creature", effects.Creature())),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.Distinct(effects.TargetCreature("another target creature"))),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Drooling Groodion — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
				}
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
					return effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Drooling Groodion — -2/-2"}.Apply(ctx)
				}
				return nil
			})),
		}},
	})
}
