package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalwart Shield-Bearers — Creature — Human Soldier {1}{W}, 0/3 (Rise of
// the Eldrazi, #5581):
//
//	"Defender
//	 Other creatures you control with defender get +0/+2."
//
// The anthem reads defender as the creature has it after layer 6, so a
// creature given defender (Guard Duty) counts and one that lost it doesn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "55aafc25-76c5-4621-941b-754bd8f13e64",
		Name:            "Stalwart Shield-Bearers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.OtherCreaturesYouControl(target, g, source) && game.HasKeyword(target, "defender")
			}, 0, 2), "Other creatures you control with defender get +0/+2."),
		},
	})
}
