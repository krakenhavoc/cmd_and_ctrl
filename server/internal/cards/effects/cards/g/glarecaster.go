package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glarecaster — Creature — Bird Cleric {4}{W}{W}, 3/3:
//
//	"Flying
//	 {5}{W}: The next time damage would be dealt to this creature and/or
//	 you this turn, that damage is dealt to any target instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection with no source,
// protecting both the Glarecaster and its controller. The ruling: damage
// dealt to both at once, or from several sources at once, is one
// instance (CR 615.8), and all of it is redirected.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c058107-b2a1-4300-8d97-c697c248df96",
		Name:         "Glarecaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.AlsoAnswers(game.AnswerRemove, effects.RedirectRow("{5}{W}: The next time damage would be dealt to this creature and/or you this turn, that damage is dealt to any target instead.",
				effects.ManaCost("{5}{W}"), effects.TargetAny(),
				effects.RedirectDamage{Protect: effects.ShieldThis, AlsoYou: true, Next: true, To: effects.RedirectToClause(0)})),
		},
	})
}
