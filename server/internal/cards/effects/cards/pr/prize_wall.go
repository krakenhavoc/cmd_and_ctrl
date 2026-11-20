package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prize Wall — Creature — Wall {1}{U}, 0/4:
//
//	"Defender
//	 {U}, {T}: You get {TK}.
//	 {4}{U}, {T}: You may put a sticker on a nonland permanent you own.
//	 Activate only as a sorcery."
//
// The ticket is a player counter (CR 107.17); the sticker is the engine's
// keyed questions (ADR 0208 §4), its ticket cost paid as it goes on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41b05346-8531-4086-96a5-e973aff150cb",
		Name:            "Prize Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: You get {TK}.",
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  effects.Do(effects.GetTickets{N: 1}),
			},
			{
				Label:        "{4}{U}, {T}: You may put a sticker on a nonland permanent you own. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{4}{U}"), effects.TapCost()),
				SorcerySpeed: true,
				Effect:       effects.Do(effects.PutASticker{}),
			},
		},
	})
}
