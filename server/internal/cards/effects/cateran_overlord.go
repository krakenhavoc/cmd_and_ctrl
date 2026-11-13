package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cateran Overlord — Creature — Horror Mercenary {4}{B}{B}{B}, 7/5
// (Mercadian Masques, #5619):
//
//	"Sacrifice a creature: Regenerate this creature.
//	 {6}, {T}: Search your library for a Mercenary permanent card with
//	 mana value 6 or less, put it onto the battlefield, then shuffle."
//
// The regeneration is Dimir House Guard's: a sacrifice-a-creature cost
// paid as the ability is activated, and a shield on the Overlord that
// replaces its next destruction this turn (CR 701.19a). Any creature
// pays it, the Overlord itself included. The search is the Ramosian
// Rebels' with the word Mercenary, at 6
// (searchForASubtypePermanentWithManaValueAtMost): the searcher may find
// nothing (CR 701.23b), and the library is shuffled either way. The {T}
// is a creature's, so the Overlord waits out summoning sickness (CR
// 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c2223cdf-419d-4153-9551-705557461664",
		Name:         "Cateran Overlord",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice a creature: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
				Cost:    SacrificeACreature(),
				Effect:  regenerateThisSource,
			},
			{
				Label:   "{6}, {T}: Search your library for a Mercenary permanent card with mana value 6 or less, put it onto the battlefield, then shuffle.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{6}"), TapCost()),
				Effect: searchForASubtypePermanentWithManaValueAtMost("Mercenary", 6,
					"Cateran Overlord — a Mercenary permanent card with mana value 6 or less, onto the battlefield"),
			},
		},
	})
}
