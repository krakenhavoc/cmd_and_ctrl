package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Legion Sergeant — Creature — Vampire Soldier {1}{B}, 2/2:
//
//	"Menace
//	 {1}{B}, Sacrifice this creature: Each nontoken Vampire creature you
//	 control gains persist until end of turn."
//
// The creatures are the ones you control as the ability resolves
// (CR 611.2c); the Sergeant itself was sacrificed to pay. Persist is
// the engine's (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8b12c43-d83a-4e5b-a525-5ee1ae850817",
		Name:            "Dusk Legion Sergeant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this creature: Each nontoken Vampire creature you control gains persist until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    Plus(ManaCost("{1}{B}"), SacrificeThis()),
			Effect: Do(GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl(), OfCreatureType("Vampire"), Not(IsTokenPredicate())),
				Keywords: []string{game.KeywordPersist},
				Label:    "Dusk Legion Sergeant — persist until end of turn",
			}),
		}},
	})
}
