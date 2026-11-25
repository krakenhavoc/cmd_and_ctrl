package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drogskol Cavalry — Creature — Spirit Knight {5}{W}{W}, 4/4 (Shadows over
// Innistrad, #5543):
//
//	"Flying
//	 Whenever another Spirit you control enters, you gain 2 life.
//	 {3}{W}: Create a 1/1 white Spirit creature token with flying."
//
// "Another" excludes the Cavalry's own entry; the Spirit may be a token, and
// is read through the layered view so a changeling counts. The ability's
// token is itself a Spirit, so it triggers the life gain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a8bac5a8-5a60-4f7a-a4d6-cb922da92472",
		Name:            "Drogskol Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Spirit"))
			}, "Drogskol Cavalry — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}: Create a 1/1 white Spirit creature token with flying.",
			Cost:    effects.ManaCost("{3}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1, CreatureTokens: 1},
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1}),
		}},
	})
}
