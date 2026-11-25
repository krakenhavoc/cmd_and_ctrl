package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crook of Condemnation — {2} Artifact (Hour of Devastation):
//
//	"{1}, {T}: Exile target card from a graveyard.
//	 {1}, Exile this artifact: Exile all graveyards."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afc8df6d-7291-400e-b768-0c0c3e3fdc32",
		Name:         "Crook of Condemnation",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}: Exile target card from a graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCardInGraveyard("target card from a graveyard"),
				Effect:  effects.ExileTheFirstLegalTarget,
			},
			{
				Label:   "{1}, Exile this artifact: Exile all graveyards.",
				Purpose: game.Purpose{Answers: game.AnswerGraveyard},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.ExileThis()),
				Effect:  effects.B02ExileAllGraveyards,
			},
		},
	})
}
