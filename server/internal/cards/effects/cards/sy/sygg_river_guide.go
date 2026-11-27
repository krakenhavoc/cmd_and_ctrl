package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sygg, River Guide — Legendary Creature — Merfolk Wizard {W}{U}, 2/2
// (Lorwyn, S82 OLD b547, #5593):
//
//	"Islandwalk
//	 {1}{W}: Target Merfolk you control gains protection from the color of
//	 your choice until end of turn."
//
// Islandwalk is the printed keyword. The ability is Benevolent Bodyguard's
// colour choice, asked as the ability resolves, on a Merfolk you control
// (Sygg is one, and so is a changeling). A target that is illegal by
// resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73f6da1e-b8ce-4b5a-9da0-84ca67adaf1c",
		Name:            "Sygg, River Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target Merfolk you control gains protection from the color of your choice until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection"))},
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target Merfolk you control", effects.OfCreatureType("Merfolk"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ProtectionFromAChosenColorForTheTarget("Sygg, River Guide")(item, effects.NewContext(g, item))
			},
		}},
	})
}
