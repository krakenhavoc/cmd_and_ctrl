package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of Zendikar — Creature — Avatar {4}{G}{G}, 6/6 (Commander 2019):
//
//	"Reach
//	 {3}{G}{G}: Create a 3/3 green Beast creature token.
//	 {3}{G}{G}, Exile this card from your graveyard: Create a 3/3 green
//	 Beast creature token."
//
// Two rows, the battlefield one and a graveyard one (Beetle, Legacy
// Criminal's shape) whose cost exiles the card, so it resolves with the
// source gone. Both make the one printed token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87217160-0687-4482-b9a3-91bad8527675",
		Name:            "Soul of Zendikar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{G}{G}: Create a 3/3 green Beast creature token.",
				Cost:    effects.ManaCost("{3}{G}{G}"),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}),
			},
			{
				Label:   "{3}{G}{G}, Exile this card from your graveyard: Create a 3/3 green Beast creature token.",
				Cost:    effects.Plus(effects.ManaCost("{3}{G}{G}"), effects.ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 green Beast"), N: 1}),
			},
		},
	})
}
