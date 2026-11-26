package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soul of Ravnica — Creature — Avatar {4}{U}{U}, 6/6 (Magic 2015, #5556):
//
//	"Flying
//	 {5}{U}{U}: Draw a card for each color among permanents you control.
//	 {5}{U}{U}, Exile this card from your graveyard: Draw a card for each
//	 color among permanents you control."
//
// Soul of Zendikar's two rows. The colors are counted as the ability resolves
// (CR 611.2c) over the permanents its controller then controls, each color
// once however many permanents have it (the effective colors, so a changed
// color counts); colorless is not a color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb014442-5935-4333-8078-ee4896dce653",
		Name:            "Soul of Ravnica",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{5}{U}{U}: Draw a card for each color among permanents you control.",
				Cost:    effects.ManaCost("{5}{U}{U}"),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.S82OldB358DrawACardForEachColorAmongYourPermanents,
			},
			{
				Label:   "{5}{U}{U}, Exile this card from your graveyard: Draw a card for each color among permanents you control.",
				Cost:    effects.Plus(effects.ManaCost("{5}{U}{U}"), effects.ExileThis()),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.S82OldB358DrawACardForEachColorAmongYourPermanents,
			},
		},
	})
}
