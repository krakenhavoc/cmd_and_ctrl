package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Compleated Huntmaster — Creature — Phyrexian Elf Warrior {2}{B}, 2/3
// (March of the Machine):
//
//	"{1}, {T}, Sacrifice another creature or artifact: Incubate 3."
//
// The tap and the sacrifice are costs, paid at announce, so the creature
// is gone before the ability resolves and a response cannot save it. The
// tap needs the Huntmaster to be free of summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65b8bb9b-91f9-44e2-aa6e-e82e3760a4d9",
		Name:         "Compleated Huntmaster",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another creature or artifact: Incubate 3.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{1}"), effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
			Effect: effects.Do(effects.Incubate{N: 3}),
		}},
	})
}
