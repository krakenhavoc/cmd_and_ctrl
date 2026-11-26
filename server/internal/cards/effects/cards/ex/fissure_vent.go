package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fissure Vent — Sorcery {3}{R}{R} (Commander 2013, S82 older sets):
//
//	"Choose one or both —
//	 • Destroy target artifact.
//	 • Destroy target nonbasic land."
//
// Run Ashore's "one or both" shape: each chosen bullet has its own target,
// the bullets resolve in printed order (CR 608.2c), and a bullet whose target
// has gone does nothing (CR 608.2b). A nonbasic land is a land without the
// basic supertype (CR 205.4c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5bac25d-72e9-4655-8a04-3646fc10be27",
		Name:         "Fissure Vent",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact()), effects.TitanOfIndustryDestroy),
			effects.ModeDoing("Destroy target nonbasic land.", effects.TargetPermanent("target nonbasic land", effects.NonbasicLand()), effects.TitanOfIndustryDestroy),
		),
	})
}
