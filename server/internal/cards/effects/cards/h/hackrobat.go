package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hackrobat — Creature — Human Rogue {1}{B}{R}, 2/3 (Ravnica Allegiance):
//
//	"Spectacle {B}{R} (You may cast this spell for its spectacle cost rather
//	 than its mana cost if an opponent lost life this turn.)
//	 {B}: This creature gains deathtouch until end of turn.
//	 {R}: This creature gets +2/-2 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "65ce903f-8c3e-4283-9b6e-ee9b3176f059",
		Name:             "Hackrobat",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Spectacle("{B}{R}")},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: This creature gains deathtouch until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("deathtouch", "Hackrobat — deathtouch until end of turn"),
			},
			{
				Label:   "{R}: This creature gets +2/-2 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(2, -2, "Hackrobat — +2/-2 until end of turn"),
			},
		},
	})
}
