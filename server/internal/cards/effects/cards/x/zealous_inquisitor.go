package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zealous Inquisitor — Creature — Human Cleric {2}{W}, 2/2:
//
//	"{1}{W}: The next 1 damage that would be dealt to this creature this
//	 turn is dealt to target creature instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// Inquisitor to the target.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e85f6251-8801-414b-adc9-5488794e7456",
		Name:         "Zealous Inquisitor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.AlsoAnswers(game.AnswerRemove, effects.RedirectRow("{1}{W}: The next 1 damage that would be dealt to this creature this turn is dealt to target creature instead.",
				effects.ManaCost("{1}{W}"), effects.TargetCreature("target creature"),
				effects.RedirectDamage{Protect: effects.ShieldThis, Amount: 1, To: effects.RedirectToClause(0)})),
		},
	})
}
