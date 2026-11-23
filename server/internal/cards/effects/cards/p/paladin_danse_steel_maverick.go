package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Paladin Danse, Steel Maverick — Legendary Artifact Creature — Synth
// Knight {2}{W}, 3/3 (Fallout, #5469):
//
//	"Vigilance, lifelink
//	 Exile Paladin Danse: Each creature you control that's an artifact
//	 or Human gains indestructible until end of turn."
//
// Hanged Executioner's exile-this cost (ExileThis) over Dora Milaje
// Elite's mass grant: the set is fixed as the ability resolves
// (CR 611.2c), so a creature that becomes an artifact or a Human later
// in the turn is not protected, and Danse, already exiled, is never in
// it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac830812-2f33-4f9e-ba3a-ea665989011c",
		Name:            "Paladin Danse, Steel Maverick",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile Paladin Danse: Each creature you control that's an artifact or Human gains indestructible until end of turn.",
			Cost:    effects.ExileThis(),
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Or(effects.Artifact(), effects.HasSubtype("Human"))),
					Keywords: []string{"indestructible"},
					Label:    "Paladin Danse, Steel Maverick — indestructible",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
