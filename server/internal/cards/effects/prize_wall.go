package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "41b05346-8531-4086-96a5-e973aff150cb",
		Name:            "Prize Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{
			{
				Label:   "{U}, {T}: You get {TK}.",
				Cost:    Plus(ManaCost("{U}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(GetTickets{N: 1}),
			},
			{
				Label:        "{4}{U}, {T}: You may put a sticker on a nonland permanent you own. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{4}{U}"), TapCost()),
				SorcerySpeed: true,
				Effect:       Do(PutASticker{}),
			},
		},
	})
}
