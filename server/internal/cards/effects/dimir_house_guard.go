package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir House Guard — Creature — Skeleton {3}{B}, 2/3 (Ravnica: City of
// Guilds, #5600):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Sacrifice a creature: Regenerate this creature.
//	 Transmute {1}{B}{B} ({1}{B}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// Fear is the printed keyword the block check reads (CR 702.36). The
// regeneration row is Wake of Vultures' with no mana, on the shared
// regenerate-this body (CR 701.19a); any creature may be sacrificed, the
// Guard itself included. Transmute is the keyword's constructor
// (transmute.go, CR 702.53a), finding a mana value 4 card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "35c027cd-f754-4e41-9242-8a9d0481171c",
		Name:            "Dimir House Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice a creature: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
				Cost:    SacrificeACreature(),
				Effect:  regenerateThisSource,
			},
			Transmute("{1}{B}{B}"),
		},
	})
}
