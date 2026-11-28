package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rathi Assassin — Creature — Phyrexian Zombie Mercenary Assassin {2}{B}{B},
// 2/2 (Nemesis, S82 OLD, #5618):
//
//	"{1}{B}{B}, {T}: Destroy target tapped nonblack creature.
//	 {3}, {T}: Search your library for a Mercenary permanent card with mana
//	 value 3 or less, put it onto the battlefield, then shuffle."
//
// Stone Catapult's removal clause on a creature's tap cost, and Rathi Fiend's
// Mercenary recruiter. Both {T} are a creature's, so summoning sickness applies
// (CR 302.6); the search may find nothing (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ea0c775-fde0-4a21-b4c2-8bcec0e4f4f8",
		Name:         "Rathi Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{B}{B}, {T}: Destroy target tapped nonblack creature.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{1}{B}{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target tapped nonblack creature", effects.And(effects.TappedPermanent(), effects.NonBlack())),
				Effect:  effects.DestroyFirstLegalTarget,
			},
			effects.NemRecruiterRow("Rathi Assassin", "{3}", "a", "Mercenary", 3),
		},
	})
}
