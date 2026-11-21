package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razia, Boros Archangel — Legendary Creature — Angel {4}{R}{R}{W}{W}, 6/3:
//
//	"Flying, vigilance, haste
//	 {T}: The next 3 damage that would be dealt to target creature you
//	 control this turn is dealt to another target creature instead."
//
// ADR 0108 §9 (#1905): a 3-point charged redirection (CR 615.7) from the
// first target to the second. The rulings: the 3 need not come from one
// source or at once; either target leaving stops it; Razia leaving does
// not.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d830a136-6fb9-42e7-81f1-97e1d713ee82",
		Name:         "Razia, Boros Archangel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.AlsoAnswers(game.AnswerRemove, effects.RedirectRow("{T}: The next 3 damage that would be dealt to target creature you control this turn is dealt to another target creature instead.",
				effects.TapCost(), effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.Distinct(effects.TargetCreature("another target creature")),
				),
				effects.RedirectDamage{Protect: effects.ShieldClause(0), Amount: 3, To: effects.RedirectToClause(1)})),
		},
	})
}
