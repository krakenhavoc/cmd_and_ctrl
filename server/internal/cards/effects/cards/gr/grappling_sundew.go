package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grappling Sundew — Creature — Plant {1}{G}, 0/4 (Guilds of Ravnica, #5525):
//
//	"Defender, reach
//	 {4}{G}: This creature gains indestructible until end of turn."
//
// Defender and reach are printed keywords; the ability is Weathered Sentinels'
// indestructible grant to the creature itself, which a Sundew that left and came
// back (a new object, CR 400.7) does not get.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a13e5a6c-e1aa-44dc-871e-120ea0390026",
		Name:            "Grappling Sundew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{G}: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{4}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"},
					Label: "Grappling Sundew — indestructible"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
