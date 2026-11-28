package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Panther — Creature — Nightstalker Cat {2}{B}, 2/2 (Mirage, S82 OLD
// b800, #5633):
//
//	"{B}, Sacrifice this creature: Destroy target creature blocking it.
//	 Sacrifice a creature named Feral Shadow, a creature named Breathstealer,
//	 and this creature: Search your library for a card named Spirit of the
//	 Night, put that card onto the battlefield, then shuffle."
//
// The first ability is Goblin Javelineer's blocker clause behind a sacrifice
// cost: the target is a creature blocking the Panther, judged from its
// last-known blocks once it has been sacrificed. The second ability's cost is
// the Panther and one permanent of each named kind (SacrificeKind.Names, a
// set rule: each named creature fills its own kind), paid at announce; the
// search is for one card of that exact name and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0678beaf-564c-4e98-b45f-a544925683a1",
		Name:         "Urborg Panther",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}, Sacrifice this creature: Destroy target creature blocking it.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.SacrificeThis()),
				Targets: effects.BlockingSource(effects.TargetCreature("target creature blocking it")),
				Effect:  effects.DestroyChosenPermanent,
			},
			{
				Label:   "Sacrifice a creature named Feral Shadow, a creature named Breathstealer, and this creature: Search your library for a card named Spirit of the Night, put that card onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Tutors: 1},
				Cost:    effects.OldB800UrborgPantherSummonCost(),
				Effect:  effects.OldB800SearchSpiritOfTheNight,
			},
		},
	})
}
