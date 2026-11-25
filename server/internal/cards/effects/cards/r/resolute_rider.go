package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Resolute Rider — Creature — Human Knight {W/B}{W/B}{W/B}{W/B}, 4/2 (Throne
// of Eldraine, S82 old sets):
//
//	"{W/B}{W/B}: This creature gains lifelink until end of turn.
//	 {W/B}{W/B}{W/B}: This creature gains indestructible until end of turn."
//
// Two activated abilities with hybrid mana costs, each granting a keyword to
// itself until end of turn, only while it is still the permanent that activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f8b2bca-991c-4dee-829b-c1cc7deaadfb",
		Name:         "Resolute Rider",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W/B}{W/B}: This creature gains lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{W/B}{W/B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("lifelink", "Resolute Rider — lifelink until end of turn"),
			},
			{
				Label:   "{W/B}{W/B}{W/B}: This creature gains indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{W/B}{W/B}{W/B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("indestructible", "Resolute Rider — indestructible until end of turn"),
			},
		},
	})
}
