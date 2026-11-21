package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "f8b12c43-d83a-4e5b-a525-5ee1ae850817",
		Name:            "Dusk Legion Sergeant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice this creature: Each nontoken Vampire creature you control gains persist until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures, Partial: true}},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeThis()),
			Effect: effects.Do(effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl(), effects.OfCreatureType("Vampire"), effects.Not(effects.IsTokenPredicate())),
				Keywords: []string{game.KeywordPersist},
				Label:    "Dusk Legion Sergeant — persist until end of turn",
			}),
		}},
	})
}
