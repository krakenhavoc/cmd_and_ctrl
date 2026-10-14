package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "65b8bb9b-91f9-44e2-aa6e-e82e3760a4d9",
		Name:         "Compleated Huntmaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice another creature or artifact: Incubate 3.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{1}"), TapCost(),
				SacrificeAnotherN(1, "another creature or artifact", Or(Creature(), Artifact()))),
			Effect: Do(Incubate{N: 3}),
		}},
	})
}
